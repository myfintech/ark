package hclutils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/myfintech/ark/src/go/lib/pattern"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/hashicorp/hcl/v2"
)

const stdinArg = "-"

/*
Large portions of this were lifted from terraform/command/fmt, but adapted to do generic HCL formatting.
There is a formatter in the HCL v1 library, but it's incompatible with HCL v2, which is what domain uses.
There is also an hclfmt program (main.go) in the HCL v2 library, but it's not very robust, nor does it
support recursive path walking.

This implementation differs from Terraform in that it takes in a pattern matcher to make the implementation
of the formatter specific to the tool using it. In the case of domain's ark CLI, it only formats BUILD.hcl and
WORKSPACE.hcl files.
*/

// FormatOpts aggregates all of the flags that are available for working with the HCL formatter
type FormatOpts struct {
	List      bool
	Write     bool
	Diff      bool
	Check     bool
	Recursive bool
	Input     io.Reader
}

// Run executes the formatting workflow
func (f *FormatOpts) Run(path string, matcher *pattern.Matcher) hcl.Diagnostics {
	diags := make(hcl.Diagnostics, 0)
	if f.Input == nil {
		f.Input = os.Stdin
	}

	if path == "" {
		path = "."
	} else if path == stdinArg {
		f.List = false
		f.Write = false
	}

	var output io.Writer
	list := f.List // preserve the original value of list
	if f.Check {
		f.List = true
		f.Write = false
		output = &bytes.Buffer{}
	} else {
		output = os.Stdout
	}

	if fmtDiags := f.fmt(path, matcher, f.Input, output); fmtDiags != nil && fmtDiags.HasErrors() {
		diags = diags.Extend(fmtDiags)
	}

	if f.Check {
		buf := output.(*bytes.Buffer)
		ok := buf.Len() == 0
		if list {
			_, _ = io.Copy(os.Stdout, buf)
		}
		if ok {
			return nil
		} else {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Input needs to be formatted",
				Detail:   "The input is not properly formatted.",
			})
		}
	}

	return diags
}

func (f *FormatOpts) fmt(path string, matcher *pattern.Matcher, reader io.Reader, output io.Writer) hcl.Diagnostics {
	diags := make(hcl.Diagnostics, 0)

	// check first to see if the path is stdin, because it's handled differently than an actual file
	if path == stdinArg {
		if f.Write {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Incorrect options passed to fmt command",
					Detail:   "Option --write cannot be used when reading from stdin.",
				},
			}
		}
		if diag := f.processFile("<stdin>", reader, output); diag != nil && diag.HasErrors() {
			return diag
		}
		return diags
	}

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Unable to get absolute path from input",
				Detail:   fmt.Sprintf("Encountered error when attempting to get an absolute path for input: %v", err),
			},
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Unable to stat provided path",
				Detail:   fmt.Sprintf("Encountered error when attempting to stat path: %v", err),
			},
		}
	}

	// check if the path is a directory because we will want to process all hcl files in the directory and optionally recurse the filesystem to find more hcl files
	if info.IsDir() {
		if dirDiags := f.processDir(path, matcher, output); dirDiags != nil && dirDiags.HasErrors() {
			return dirDiags
		}
		return diags
	}

	// check if the path is an hcl file; if it's not, we can't/shouldn't process it
	if filepath.Ext(path) != ".hcl" {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Non HCL file provided",
				Detail:   "Only files with the '.hcl' file extension are eligible for formatting.",
			},
		}
	}

	// check if hcl file matches a provided set of match patterns
	if matcher == nil || matcher.Check(path) {
		log.Debugf("formatting file: %s", path)
		file, openErr := os.Open(path)
		defer func() {
			_ = file.Close()
		}()
		if openErr != nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Unable to open file",
					Detail:   fmt.Sprintf("Encountered errer when attempted to open file: %v", err),
				},
			}
		}
		if diag := f.processFile(path, file, output); diag != nil && diag.HasErrors() {
			return diag
		}
		return diags
	}

	// send an error if an explicit HCL file is passed to the tool that does not meet the tool's inclusion patterns
	if !matcher.Check(path) {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Provided HCL file does not match against tool's configuration file patterns",
				Detail:   fmt.Sprintf("Tool match patterns are this: %s; provided file is this: %s", strings.Join(matcher.Includes, ", "), path),
			},
		}
	}

	return diags
}

func (f *FormatOpts) processFile(path string, reader io.Reader, writer io.Writer) hcl.Diagnostics {
	var info os.FileInfo

	if path != "<stdin>" {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Could not get OS info for provided path",
					Detail:   "Attempted to get OS information for provided path, but failed to do so.",
				},
			}
		}
		info = fileInfo
	}

	src, err := ioutil.ReadAll(reader)
	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Unable to read input",
				Detail:   "Attempted to read from provided IO reader and failed",
			},
		}
	}

	if _, diag := hclsyntax.ParseConfig(src, path, hcl.Pos{Line: 1, Column: 1}); diag != nil && diag.HasErrors() {
		return diag
	}

	result := hclwrite.Format(src)

	if !bytes.Equal(src, result) {
		log.Debugf("hcl fmt: Formatting %s", path)
		if f.List {
			_, _ = fmt.Fprintln(writer, path)
		}
		if f.Write {
			info, err = os.Stat(path)
			if err != nil {
				return hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "Unable to stat file",
						Detail:   fmt.Sprintf("Encountered error when attempting to stat %s: %v", path, err),
					},
				}
			}
			if writeErr := ioutil.WriteFile(path, result, info.Mode()); writeErr != nil {
				return hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "Unable to write formatted file",
						Detail:   "Attempted to write formatted file and failed.",
					},
				}
			}

		}
		if f.Diff {
			diff, diffErr := bytesDiff(src, result, path)
			if diffErr != nil {
				return hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "Unable to diff source and result",
						Detail:   "Attempted to run diff on source data and formatted data and failed.",
					},
				}
			}
			_, _ = writer.Write(diff)
		}
	}

	if !f.List && !f.Write && f.Diff {
		if _, err = writer.Write(result); err != nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Failed to write result",
					Detail:   "Attempted to write result and failed.",
				},
			}
		}
	}

	return nil
}

func (f *FormatOpts) processDir(startPath string, matcher *pattern.Matcher, output io.Writer) hcl.Diagnostics {
	diags := make(hcl.Diagnostics, 0)
	log.Debugf("hcl fmt: Looking for HCL files in %s", startPath)
	if f.Recursive {
		err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return errors.Wrapf(err, "unable to access path: %s", path)
			}
			if filepath.Ext(info.Name()) == ".hcl" {
				if matcher == nil || matcher.Check(path) {
					log.Debugf("formatting file: %s", path)
					fileHandle, openErr := os.Open(path)
					if openErr != nil {
						return errors.Wrapf(openErr, "there was an issue opening file: %s", path)
					}
					processDiags := f.processFile(path, fileHandle, output)
					diags = diags.Extend(processDiags)
				}
			}
			return nil
		})
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "There was a problem walking the filsystem",
				Detail:   err.Error(),
			})
		}
	} else {
		entries, err := ioutil.ReadDir(startPath)
		if os.IsNotExist(err) {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Directory not found",
				Detail:   fmt.Sprintf("There is no directory at %s", startPath),
			})
		}
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to read directory",
				Detail:   fmt.Sprintf("Unable to read directory at %s", startPath),
			})
		}

		for _, info := range entries {
			name := info.Name()
			subPath := filepath.Join(startPath, name)
			if filepath.Ext(name) == ".hcl" {
				if matcher == nil || matcher.Check(subPath) {
					log.Debugf("formatting file: %s", subPath)
					file, openErr := os.Open(subPath)
					if openErr != nil {
						diags = diags.Append(&hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to open file",
							Detail:   fmt.Sprintf("Unable to read file: %s", subPath),
						})
						continue
					}
					if processDiags := f.processFile(subPath, file, output); processDiags != nil && processDiags.HasErrors() {
						diags = diags.Extend(processDiags)
						continue
					}
				}
			}
		}
	}

	return diags
}

func bytesDiff(b1, b2 []byte, path string) (data []byte, err error) {
	f1, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer func() {
		_ = os.Remove(f1.Name())
	}()
	defer func() {
		_ = f1.Close()
	}()

	f2, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer func() {
		_ = os.Remove(f2.Name())
	}()
	defer func() {
		_ = f2.Close()
	}()

	_, _ = f1.Write(b1)
	_, _ = f2.Write(b2)

	data, err = exec.Command("diff", "--label=old/"+path, "--label=new/"+path, "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return
}
