package base

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

// Targets is a map of types to their target interfaces
// Targets["exec"]emptyExampleExecTarget{}
type Targets map[string]Buildable

// Register adds a target type to the map
func (ttm Targets) Register(targetType string, target Buildable) {
	ttm[targetType] = target
}

/*
MapRawTargets reflects against an internal map of buildable target types
constructs a new copy of the target type in its memory, decodes the remaining rawTarget HCL
into the new copy of the mapped target and embeds the rawTarget in the mappedTarget

if there is no target type for the rawTarget.Type an error will be raised
if the mapped target cannot be converted into a pointer an error will be raised
if the mapped target cannot be cast back into the target interface an error will be raised.
*/
func (ttm Targets) MapRawTargets(rawTargets []*RawTarget) ([]Buildable, hcl.Diagnostics) {
	var targets []Buildable
	for _, rawTarget := range rawTargets {

		// lookup a targetInterface for the declared rawTarget.Type
		targetInterface, exists := ttm[rawTarget.Type]

		// no targetInterface was register in the map
		// we won't be able to reflect on this
		if !exists {
			return targets, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("missing target instance for type %s", rawTarget.Type),
					Detail:   fmt.Sprintf("missing target instance for type %s", rawTarget.Type),
				},
			}
		}

		// create a new instance of the target type registered in the map
		targetVal := reflect.New(reflect.TypeOf(targetInterface))

		// attempt to decode the remaining rawTarget.HCL body into the reflected target
		diag := gohcl.DecodeBody(rawTarget.HCL, nil, targetVal.Interface())
		if diag != nil && diag.HasErrors() {
			return targets, diag
		}

		// inject RawTarget into the newly constructed target interface
		rtField := targetVal.Elem().FieldByName("RawTarget")
		if rtField.IsValid() {
			rtField.Set(reflect.ValueOf(rawTarget))
		}

		// cast the reflected target into the target interface
		target, ok := targetVal.Elem().Interface().(Buildable)
		if !ok {
			return targets, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("cannot reflect target type %s", rawTarget.Type),
					Detail:   fmt.Sprintf("cannot reflect target type %s", rawTarget.Type),
				},
			}
		}

		// append the final target into the targets slice
		targets = append(targets, target)
	}
	return targets, nil
}
