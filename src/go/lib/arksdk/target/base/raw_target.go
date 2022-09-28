package base

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/myfintech/ark/src/go/lib/pattern"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/fs/observer"
	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/state_store"
	"github.com/myfintech/ark/src/go/lib/utils/cloudutils"
	"github.com/myfintech/ark/src/go/lib/utils/cryptoutils"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

const (
	attrsStateKey = "attributes"
)

// Package the package of the give target
// This is used to build target addressing
type Package struct {
	Name        string `hcl:",label"`
	Description string `hcl:"description,attr"`
	Version     string `hcl:"version,optional"`
	Dir         string
}

// Describe returns a string description of a package
func (p Package) Describe() string {
	return p.Description
}

// Locals represents variables that are constrained to the current package/BUILD.hcl file
type Locals struct {
	HCL    hcl.Body `json:"-" hcl:",remain"`
	Values map[string]cty.Value
}

// Module defines the properties used to create re-usable modules
type Module struct {
	Name   string         `hcl:",label"`
	Source hcl.Expression `hcl:"source,attr"`
	HCL    hcl.Body       `json:"-" hcl:",remain"`
	Path   string
}

// EvalAttributes evaluates the attributes from the given HCL evaluation context
func (m *Module) EvalAttributes(ctx *hcl.EvalContext) (map[string]cty.Value, hcl.Diagnostics) {
	vars := make(map[string]cty.Value)
	attributes, hclDiag := m.HCL.JustAttributes()

	if hclDiag != nil && hclDiag.HasErrors() {
		return vars, hclDiag
	}
	for _, v := range attributes {
		ctyVal, rangeDiag := v.Expr.Value(ctx)
		if rangeDiag != nil && rangeDiag.HasErrors() {
			return vars, rangeDiag
		}
		vars[v.Name] = ctyVal
	}
	return vars, nil
}

// RawTarget a data container for the initial decoding pass
type RawTarget struct {
	// initial pass properties of a target
	Type              string    `hcl:",label"`
	Name              string    `hcl:",label"`
	Description       *string   `hcl:"description,attr"`
	DependsOn         *[]string `hcl:"depends_on,attr"`
	DeclaredArtifacts *[]string `hcl:"artifacts,attr"`

	Labels hcl.Expression `hcl:"labels,optional"`

	SourceFiles     hcl.Expression `hcl:"source_files,attr"`
	IncludePatterns hcl.Expression `hcl:"include_patterns,attr"`
	ExcludePatterns hcl.Expression `hcl:"exclude_patterns,attr"`

	// A file matcher used for filtering source files
	FileMatcher *pattern.Matcher `json:"-"`

	// the remaining HCL encoded body
	HCL      hcl.Body `json:"-" hcl:",remain"`
	HCLBytes []byte   `json:"-"`

	// the directory of the BUILD.hcl file
	Dir string

	// the BUILD.hcl file this came from
	File string

	Package   *Package   `json:"-"`
	Workspace *Workspace `json:"-"`
	Locals    *Locals    `json:"-"`

	// State Management
	StateStore *state_store.StateStore

	RawComputedAttrs *RawComputedAttrs
	Module           *Module
}

// RawComputedAttrs is a container for hcl expressions that need to be referenced
type RawComputedAttrs struct {
	Labels          []string `hcl:"labels,optional"`
	Files           []string `hcl:"source_files,optional"`
	IncludePatterns []string `hcl:"include_patterns,optional"`
	ExcludePatterns []string `hcl:"exclude_patterns,optional"`
}

// IsolateHCLBlocks uses the targets source bytes to isolate its original HCL form (convenient for rewriting or hashing)
func (t RawTarget) IsolateHCLBlocks() (*hclwrite.File, error) {
	return hclutils.IsolateBlocks(t.HCLBytes, func(file *hclwrite.File) *hclwrite.Block {
		return file.Body().FirstMatchingBlock("target", []string{t.Type, t.Name})
	}, t.File)
}

// GetRawTarget returns the RawTarget
func (t *RawTarget) GetRawTarget() *RawTarget {
	return t
}

// EvaluateHCLExpressions evaluates the source_files, include_patterns, and exclude patterns of the target
func (t *RawTarget) EvaluateHCLExpressions() error {
	expressions := RawComputedAttrs{}

	err := hclutils.DecodeExpressions(t, &expressions, CreateEvalContext(EvalContextOptions{
		TargetLookupTable:            t.Workspace.TargetLUT,
		Package:                      *t.Package,
		Workspace:                    *t.Workspace,
		DisableLookupTableEvaluation: true,
	}))

	if err != nil {
		return err
	}

	t.RawComputedAttrs = &expressions

	for i, file := range expressions.Files {
		if !filepath.IsAbs(file) {
			expressions.Files[i] = filepath.Clean(filepath.Join(t.Dir, file))
		}
	}

	t.FileMatcher = &pattern.Matcher{
		Paths:    expressions.Files,
		Includes: expressions.IncludePatterns,
		Excludes: expressions.ExcludePatterns,
	}

	if err = t.Workspace.Observer.AddFileMatcher(t.Address(), t.FileMatcher); err != nil {
		return err
	}

	return nil
}

// PrevState return the previous state from the state store
func (t *RawTarget) PrevState() state_store.KVStore {
	return t.StateStore.Previous
}

// State return the current state from the state store
func (t *RawTarget) State() state_store.KVStore {
	return t.StateStore.Current
}

// GetStateAttrs returns the stored computed attrs
func (t *RawTarget) GetStateAttrs() interface{} {
	return t.State().Get(attrsStateKey)
}

// SetStateAttrs updates the attrsStateKey in the state store
func (t *RawTarget) SetStateAttrs(v interface{}) interface{} {
	return t.State().Set(attrsStateKey, v)
}

// Deps returns the rawTarget.DependsOn
func (t *RawTarget) Deps() []string {
	if t.DependsOn == nil {
		return []string{}
	}
	return *t.DependsOn
}

// Address returns a string that can be used to address this target
func (t RawTarget) Address() string {
	return BuildAddress(t.Package.Name, t.Type, t.Name)
}

// Describe returns a string description of a target
func (t RawTarget) Describe() string {
	if t.Description == nil {
		return ""
	}
	return *t.Description
}

// ListLabels returns a string slice of labels on a target
func (t RawTarget) ListLabels() []string {
	return t.RawComputedAttrs.Labels
}

// Hashcode an interface used as a key by the DAG lib
func (t RawTarget) Hashcode() interface{} {
	return t.Address()
}

// String an alias for Address
func (t RawTarget) String() string {
	return t.Address()
}

// DirRelToWorkspace the target's directory relative to the workspace root
func (t RawTarget) DirRelToWorkspace() string {
	return strings.TrimPrefix(strings.TrimPrefix(t.Dir, t.Workspace.Dir), "/")
}

// ArtifactsDir returns the file path to the artifacts directory
func (t RawTarget) ArtifactsDir() string {
	return filepath.Join(t.Workspace.ArtifactsDir(), t.Address(), t.Hash())
}

// TmpArtifactsDir returns an artifacts directory in the os.TempDir()
func (t RawTarget) TmpArtifactsDir() string {
	return filepath.Join(os.TempDir(), t.DirRelToWorkspace(), t.Type, t.Name)
}

// MkArtifactsDir creates the artifacts directory
func (t RawTarget) MkArtifactsDir() error {
	return os.MkdirAll(t.ArtifactsDir(), 0755)
}

// MkTmpArtifactsDir creates the artifacts directory in os.TempDir()
func (t RawTarget) MkTmpArtifactsDir() error {
	return os.MkdirAll(t.TmpArtifactsDir(), 0755)
}

// Hash returns a hex formatted sha256 hash of the sources and dependencies
// If the hash is a zero hash we panic
func (t RawTarget) Hash() string {
	if hash := t.State().GetString("hash"); hash != "" {
		return hash
	}

	rootHash := sha256.New()
	vertices := t.Workspace.TargetGraph.Isolate(t).TopologicalSort(t)

	for _, vertex := range vertices {
		isolator := vertex.(hclutils.Isolator)
		sourceHCL, _ := isolator.IsolateHCLBlocks()
		_, _ = hclutils.HashFile(sourceHCL, rootHash)

		addressable := vertex.(Addressable)
		if fileCache, ok := t.Workspace.Observer.GetMatchCache(addressable.Address()); ok && fileCache != nil {
			_, _ = fmt.Fprintf(rootHash, "%x\n", fileCache.Hash.Sum(nil))
		}
	}

	finalHash := hex.EncodeToString(rootHash.Sum(nil))

	ctxLog.
		WithField("target", t.Address()).
		WithField("hash", finalHash[:7]).
		WithField("is_zero", cryptoutils.CompareHashes(sha256.New(), rootHash)).
		Trace("computed hash")

	t.State().Set("hash", finalHash)
	return finalHash
}

// FileDeps returns an observer file match cache
func (t RawTarget) FileDeps() (*observer.FileMatchCache, bool) {
	return t.Workspace.Observer.GetMatchCache(t.Address())
}

// SourceFilesList returns a string slice of source files
func (t RawTarget) SourceFilesList() (sourceFiles []string) {
	if mc, _ := t.Workspace.Observer.GetMatchCache(t.Address()); mc != nil {
		return mc.FilesStringList(nil)
	}
	return
}

// ShortHash the first 7 characters of the calculated hash of the sources and dependencies
func (t RawTarget) ShortHash() string {
	rootHash := t.Hash()
	if len(rootHash) < 7 {
		return rootHash
	}
	return rootHash[:7]
}

// Artifacts a calculated list of artifacts that were declared
func (t RawTarget) Artifacts() map[string]string {
	artifacts := map[string]string{
		"path": t.ArtifactsDir(),
	}

	if t.DeclaredArtifacts == nil {
		return artifacts
	}

	for _, artifact := range *t.DeclaredArtifacts {
		artifacts[artifact] = filepath.Join(t.ArtifactsDir(), artifact)
	}

	return artifacts
}

// LocalVars returns a map of variables local to the package for use in hcl.EvalContexts
func (t RawTarget) LocalVars() map[string]cty.Value {
	if t.Locals != nil {
		return t.Locals.Values
	}
	return make(map[string]cty.Value)
}

// Attributes a map of attributes for use in hcl.EvalContexts
func (t RawTarget) Attributes() map[string]cty.Value {
	return map[string]cty.Value{
		"name":         cty.StringVal(t.Name),
		"type":         cty.StringVal(t.Type),
		"hash":         cty.StringVal(t.Hash()),
		"short_hash":   cty.StringVal(t.ShortHash()),
		"dir":          cty.StringVal(t.Dir),
		"path":         cty.StringVal(t.Dir),
		"address":      cty.StringVal(t.Address()),
		"dir_rel":      cty.StringVal(t.DirRelToWorkspace()),
		"artifacts":    hclutils.MapStringStringToCtyObject(t.Artifacts()),
		"source_files": hclutils.StringSliceToCtyTuple(t.SourceFilesList()),
	}
}

// DecodedBuildFile ...
type DecodedBuildFile struct {
	Package    *Package     `hcl:"package,block"`
	Locals     *Locals      `hcl:"locals,block"`
	RawTargets []*RawTarget `hcl:"target,block"`
	Modules    []*Module    `hcl:"module,block"`
}

// DecodeRawTargetOpts is a struct of options to pass to DecodeRawTargetsFromHCLFile()
type DecodeRawTargetOpts struct {
	HclFile          *hcl.File
	Filename         string
	Workspace        *Workspace
	HclCtx           *hcl.EvalContext
	DecodedBuildFile *DecodedBuildFile
	MaxDepth         int
	CurrentDepth     int
}

// DecodeRawTargetsFromHCLFile returns a slice of RawTarget pointers after the initial decoding pass
func DecodeRawTargetsFromHCLFile(opts DecodeRawTargetOpts) ([]*RawTarget, hcl.Diagnostics) {
	opts.CurrentDepth++

	diag := gohcl.DecodeBody(opts.HclFile.Body, opts.HclCtx, opts.DecodedBuildFile)

	if diag != nil && diag.HasErrors() {
		return opts.DecodedBuildFile.RawTargets, diag
	}

	if opts.DecodedBuildFile.Package == nil {
		return opts.DecodedBuildFile.RawTargets, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "the decoded build file's package field cannot be nil",
				Detail:      fmt.Sprintf("tha package field for %v was nil", opts.DecodedBuildFile),
				EvalContext: opts.HclCtx,
			},
		}
	}

	opts.DecodedBuildFile.Package.Dir = filepath.Dir(opts.Filename)

	// FIXME: locals that call locals (that call locals ...) might not be evaluated in the correct order, resulting in errors
	if opts.DecodedBuildFile.Locals != nil {
		opts.DecodedBuildFile.Locals.Values = make(map[string]cty.Value)
		attributes, localDiag := opts.DecodedBuildFile.Locals.HCL.JustAttributes()
		if localDiag != nil && localDiag.HasErrors() {
			return opts.DecodedBuildFile.RawTargets, localDiag
		}
		sortedAttributes := hclutils.SortAttributes(attributes)
		for _, v := range sortedAttributes {
			ctyVal, rangeDiag := v.Expr.Value(&hcl.EvalContext{
				Functions: hclutils.BuildStdLibFunctions(opts.Workspace.Dir),
				Variables: map[string]cty.Value{
					"locals": cty.ObjectVal(opts.DecodedBuildFile.Locals.Values),
				},
			})
			if rangeDiag != nil && rangeDiag.HasErrors() {
				return opts.DecodedBuildFile.RawTargets, rangeDiag
			}
			opts.DecodedBuildFile.Locals.Values[v.Name] = ctyVal
		}
	}

	for _, rawTarget := range opts.DecodedBuildFile.RawTargets {
		rawTarget.File = opts.Filename
		rawTarget.HCLBytes = opts.HclFile.Bytes
		rawTarget.Dir = opts.DecodedBuildFile.Package.Dir
		rawTarget.Workspace = opts.Workspace
		rawTarget.Package = opts.DecodedBuildFile.Package
		rawTarget.Locals = opts.DecodedBuildFile.Locals
		rawTarget.StateStore = state_store.NewStateStore()
	}

	for _, module := range opts.DecodedBuildFile.Modules {
		if opts.MaxDepth > 0 && opts.CurrentDepth > opts.MaxDepth {
			return opts.DecodedBuildFile.RawTargets, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "a cyclic reference was detected in your module",
					Detail:      fmt.Sprintf("max depth %d, current depth %d %s", opts.MaxDepth, opts.CurrentDepth, module.Source),
					EvalContext: opts.HclCtx,
				},
			}
		}

		modulePath, modDiag := module.Source.Value(opts.HclCtx)
		if modDiag != nil && modDiag.HasErrors() {
			return opts.DecodedBuildFile.RawTargets, modDiag
		}
		// This should not be able to be nil as we're checking the hcl diag for a nil error above
		module.Path = modulePath.AsString()

		rawHCLFile, rawDiag := hclutils.FileFromPath(module.Path)
		if rawDiag != nil && rawDiag.HasErrors() {
			return opts.DecodedBuildFile.RawTargets, rawDiag
		}

		modOpts := DecodeRawTargetOpts{
			HclFile:  rawHCLFile,
			Filename: module.Path,
			DecodedBuildFile: &DecodedBuildFile{
				Package: &Package{
					Name:        fmt.Sprintf(opts.DecodedBuildFile.Package.Name + "." + module.Name),
					Description: "",
					Version:     "",
				},
			},
			Workspace: opts.Workspace,
			HclCtx:    opts.HclCtx,
			MaxDepth:  100,
		}

		targets, targetDiag := DecodeRawTargetsFromHCLFile(modOpts)
		if targetDiag != nil && targetDiag.HasErrors() {
			return opts.DecodedBuildFile.RawTargets, targetDiag
		}
		for _, target := range targets {
			target.Module = module
		}
		opts.DecodedBuildFile.RawTargets = append(opts.DecodedBuildFile.RawTargets, targets...)
	}
	return opts.DecodedBuildFile.RawTargets, nil
}

// StateFilePath returns a string path to the state file for the target
func (t RawTarget) StateFilePath() string {
	return filepath.Join(t.ArtifactsDir(), "state.json")
}

// CheckLocalBuildCache loads the build cache state and verifies that the hashes match
func (t RawTarget) CheckLocalBuildCache() (bool, error) {
	log.Debugf("Checking local artifact %s", t.ArtifactsDir())
	state, err := t.LoadLocalBuildCacheState()
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return state.Hash == t.Hash(), nil
}

// LoadLocalBuildCacheState decodes the local build cache state file for the specified target
func (t RawTarget) LoadLocalBuildCacheState() (BuildCacheState, error) {
	state := BuildCacheState{}
	stateFile, err := ioutil.ReadFile(t.StateFilePath())
	if err != nil {
		return state, err
	}

	if err = json.Unmarshal(stateFile, &state); err != nil {
		return state, err
	}

	return state, nil
}

// SaveLocalBuildCacheState saves a new instance of BuildCacheState as JSON to t.StateFilePath()
func (t RawTarget) SaveLocalBuildCacheState() error {
	state := BuildCacheState{
		Name: t.Name,
		Type: t.Type,
		Hash: t.Hash(),
	}

	stateBytes, err := json.MarshalIndent(state, "", "\t")
	if err != nil {
		return err
	}

	stateFileDir := filepath.Dir(t.StateFilePath())
	if err = os.MkdirAll(stateFileDir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(t.StateFilePath(), stateBytes, 0644)
}

// RemoteArtifactKey constructs a string path for a remote artifact
func (t RawTarget) RemoteArtifactKey() string {
	return filepath.Join("artifacts", t.Address(), fmt.Sprintf("%s.tar.gz", t.Hash()))
}

// CheckRemoteCache determines whether an artifact for a given target exists in remote storage
func (t RawTarget) CheckRemoteCache() (bool, error) {
	return cloudutils.BlobCheck(nil, t.Workspace.Config.Artifacts.StorageBaseURL, t.RemoteArtifactKey())
}

// PullRemoteCache pulls an artifact from remote storage that matches a target's state hash and loads it for local use
func (t RawTarget) PullRemoteCache() error {
	reader, cleanup, pullErr := cloudutils.NewBlobReader(nil, t.Workspace.Config.Artifacts.StorageBaseURL, t.RemoteArtifactKey())
	if pullErr != nil {
		return errors.Wrap(pullErr, "unable to pull artifact from remote cache")
	}
	defer cleanup()

	if err := fs.GzipUntar(t.ArtifactsDir(), reader); err != nil {
		return errors.Wrap(err, "unable to expand compressed artifact into destination directory")
	}

	return nil
}

// PushRemoteCache archives and compresses an artifact and pushes the artifact into remote storage
func (t RawTarget) PushRemoteCache() error {
	writer, cleanup, pushErr := cloudutils.NewBlobWriter(nil, t.Workspace.Config.Artifacts.StorageBaseURL, t.RemoteArtifactKey())
	if pushErr != nil {
		return errors.Wrap(pushErr, "unable to create a writer for the remote destination")
	}
	defer cleanup()

	if err := fs.GzipTar(t.ArtifactsDir(), writer); err != nil {
		return errors.Wrap(err, "unable to archive artifact to remote destination")
	}

	return nil
}

// NewRawTarget initializes a raw target with a pointer to a workspace, and an empty state store

// GetFileMatcher returns the FileMatcher on the RawTarget used in the Buildable interface
func (t RawTarget) GetFileMatcher() *pattern.Matcher {
	return t.FileMatcher
}

// PackageName returns the string value of package.Name
func (t RawTarget) PackageName() string {
	if t.Package != nil {
		return t.Package.Name
	}
	return ""
}

// PackageVersion returns the string value of package.Version
func (t RawTarget) PackageVersion() string {
	if t.Package != nil {
		return t.Package.Version
	}
	return ""
}

// PackageDir is the absolute path of the directory of a target
func (t RawTarget) PackageDir() string {
	return t.Dir
}

// HashableAttributes returns a map of attributes that are hashed for the purpose of tracking build configuration changes for a target
// Deprecated
func (t RawTarget) HashableAttributes() map[string]interface{} {
	// attrs, _ := t.HCL.JustAttributes()

	return map[string]interface{}{
		// TODO: figure out HCL body
		"name":       t.Name,
		"type":       t.Type,
		"depends_on": t.DependsOn,
		// "declared_artifacts": t.DeclaredArtifacts,
		"package": t.Package,
		// "attributes":         attrs,
	}
}

// CacheEnabled returns true to cache targets by default. Can be overridden on a per-target basis
func (t RawTarget) CacheEnabled() bool {
	return true
}
