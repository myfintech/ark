package base

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/myfintech/ark/src/go/lib/pattern"

	"github.com/hashicorp/hcl/v2"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/myfintech/ark/src/go/lib/utils"
)

// LookupTable (look up table) adds and retrieves targets by their target.Address()
// LookupTable["src/go/tools/monarch.go_binary:linux"]Addressable
type LookupTable map[string]Addressable

// BuildAddress constructs the target address string from a directory, type, and name
func BuildAddress(packageName, targetType, targetName string) string {
	return fmt.Sprintf("%s.%s.%s", packageName, targetType, targetName)
}

// ParsedAddress a struct that represents the parts of a parsed address
type ParsedAddress struct {
	Package string
	Type    string
	Name    string
}

// String returns a string representation of an address
func (p ParsedAddress) String() string {
	return strings.Join([]string{p.Package, p.Type, p.Name}, ".")
}

// ParseAddress parses an address string and returns an parsed address struct
func ParseAddress(address string) (*ParsedAddress, error) {
	parts := strings.Split(address, ".")

	if len(parts) < 3 {
		return nil, errors.Errorf("address had an invalid length of %d expected 3 parts %s", len(parts), address)
	}

	return &ParsedAddress{
		Package: parts[0],
		Type:    parts[1],
		Name:    parts[2],
	}, nil
}

var reservedPackageNames = map[string]bool{
	"workspace": true,
}

// Add adds a target to the LUT
func (t LookupTable) Add(target Addressable) error {
	parsedAddress, err := ParseAddress(target.Address())
	if err != nil {
		return err
	}
	if _, exists := reservedPackageNames[parsedAddress.Package]; exists {
		return errors.Errorf("package name '%s' is a reserved name that cannot be implemented; please choose a different package name", parsedAddress.Package)
	}
	if _, exists := t[target.Address()]; exists {
		return errors.Errorf("target '%s' already exists; targets must be unique across the entire workspace", target.Address())
	}
	t[target.Address()] = target
	return nil
}

// Lookup locates a target by constructing its address
// This is a convenience function to be used as an hcl eval function and the CLI
func (t LookupTable) Lookup(packageName, targetType, targetName string) (Addressable, error) {
	return t.LookupByAddress(BuildAddress(packageName, targetType, targetName))
}

// LookupByAddress uses a fully qualified address
func (t LookupTable) LookupByAddress(address string) (Addressable, error) {
	if target, exists := t[address]; exists {
		return target, nil
	}

	return nil, errors.Errorf("failed to lookup target by address %s", address)
}

// SortedAddresses returns an alphabetical list of addresses
func (t LookupTable) SortedAddresses() []string {
	var addresses []string
	for address := range t {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)
	return addresses
}

// FilterByPatterns returns a slice of Addressable targets that have been filtered against a slice of glob patterns supplied by the pattern.Matcher
func (t LookupTable) FilterByPatterns(matcher *pattern.Matcher) (addressables []Addressable) {
	for _, addressable := range t {
		if matcher == nil {
			addressables = append(addressables, addressable)
			continue
		}
		for _, label := range addressable.ListLabels() {
			if matcher.Included(label) {
				addressables = append(addressables, addressable)
			}
		}
	}
	return
}

// FilterSortedAddresses returns a sorted slice of target addresses that have been filtered against a slice of glob patterns supplied by the pattern.Matcher
func (t LookupTable) FilterSortedAddresses(matcher *pattern.Matcher) (addresses []string) {
	for _, addressable := range t.FilterByPatterns(matcher) {
		addresses = append(addresses, addressable.Address())
	}
	sort.Strings(addresses)
	return
}

// ToCtyVariables returns the lookup table as a deeply nested object for use in hcl.Expressions
func (t LookupTable) ToCtyVariables() map[string]cty.Value {
	packages := make(map[string]interface{})

	for _, addressable := range t {
		if target, ok := addressable.(Target); ok {
			utils.BuildDeepMapString(target.Attributes(), strings.Split(target.Address(), "."), packages)
		} else {
			log.Debugf("cant compile variables for on %s as it does not implement the Target interface", addressable.Address())
		}
	}

	return hclutils.MapStringInterfaceToCty(packages)
}

// ResolveHCLVariableDependencies returns a list of addressables from the lookup table that are referenced as variables in hcl.Expressions
// This function will panic if you pass a non struct type as it uses reflection to look for HCL expression struct fields
func (t LookupTable) ResolveHCLVariableDependencies(addressable Addressable) []Addressable {
	var variableDeps []Addressable

	targetVal := reflect.ValueOf(addressable)

	// iterate over all target fields
	for i := 0; i < targetVal.NumField(); i++ {
		field := targetVal.Field(i)
		// extract the expression from the hcl.Expression field.
		if exp, ok := field.Interface().(hcl.Expression); ok {
			// iterate over the variables in the expression (hcl.Traversals)
			for _, traversal := range exp.Variables() {
				// build a set of keys for every variable (traversal)
				// [package_name, target_type, target_name]
				parsedAddress, err := ParseAddress(hclutils.TraversalToString(traversal))

				// not a valid address
				if err != nil {
					continue
				}

				// attempt to locate the target in the lookup table
				if dep, exists := t[parsedAddress.String()]; exists {
					// we've located a target add it to the dependencies list
					variableDeps = append(variableDeps, dep)
				}
			}
		}
	}
	return variableDeps
}

// BuildGraph returns a graph of all objects in the lookup table
func (t LookupTable) BuildGraph() (*Graph, error) {
	graph := new(Graph)

	for _, addressable := range t {
		if err := t.BuildTargetDependencies(graph, addressable); err != nil {
			return graph, err
		}
	}

	return graph, nil
}

// BuildTargetDependencies takes a graph and an addressable and attempts to build build onto the graph
func (t *LookupTable) BuildTargetDependencies(graph *Graph, addressable Addressable) error {
	target, ok := addressable.(Target)
	if !ok {
		return errors.Errorf("the resource at address: %s doesnt implement the target interface", addressable.Address())
	}
	return t.BuildGraphEdges(graph, target)
}

// BuildGraphEdges takes a graph and a target and attempts to build edges between the target and its dependencies
func (t LookupTable) BuildGraphEdges(graph *Graph, target Target) error {
	graph.Add(target)

	// resolve dependencies referenced from variables in HCL
	for _, dep := range t.ResolveHCLVariableDependencies(target) {
		graph.Add(dep)
		graph.Connect(target, dep)
	}
	// resolve dependencies explicitly declared in HCL
	for _, depAddress := range target.Deps() {
		dep, exists := t[depAddress]
		if !exists {
			// TODO: raise diagnostic error instead of raw error to include details of what file the error is in
			return errors.Errorf("failed to locate dependency: %s for target: %s in the lookup table", depAddress, target.Address())
		}
		// connect the target, and it's dep
		graph.Add(dep)
		graph.Connect(target, dep)
	}
	return nil
}
