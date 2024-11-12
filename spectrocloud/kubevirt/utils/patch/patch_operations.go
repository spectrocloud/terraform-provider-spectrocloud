package patch

import (
	"reflect"
	"slices"
	"strings"
)

var patchKeyEncoder = strings.NewReplacer("~", "~0", "/", "~1")

const (
	opAdd     = "add"
	opRemove  = "remove"
	opReplace = "replace"
)

type Operation struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
	Op    string      `json:"op"`
}

func DiffStringMap(pathPrefix string, oldV, newV map[string]interface{}) Operations {
	pathPrefix = strings.TrimRight(pathPrefix, "/")
	// If old value was empty, just create the object
	if len(oldV) == 0 {
		return Operations{
			{
				Path:  pathPrefix,
				Value: newV,
				Op:    opAdd,
			},
		}
	}

	// This is suboptimal for adding whole new map from scratch
	// or deleting the whole map, but it's actually intention.
	// There may be some other map items managed outside of TF
	// and we don't want to touch these.
	var ops Operations
	for k := range oldV {
		if _, ok := newV[k]; ok {
			continue
		}
		ops = append(ops, Operation{
			Path: pathPrefix + "/" + patchKeyEncoder.Replace(k),
			Op:   opRemove,
		})
	}

	for k, v := range newV {
		newValue := v.(string)

		if oldValue, ok := oldV[k].(string); ok {
			if oldValue != newValue {
				ops = append(ops, Operation{
					Path:  pathPrefix + "/" + patchKeyEncoder.Replace(k),
					Value: newValue,
					Op:    opReplace,
				})
			}
		} else {
			ops = append(ops, Operation{
				Path:  pathPrefix + "/" + patchKeyEncoder.Replace(k),
				Value: newValue,
				Op:    opAdd,
			})
		}
	}
	return ops
}

type Operations []Operation

func (o Operations) Equal(ops Operations) bool {
	for _, op := range o {
		if !slices.ContainsFunc(ops, func(operation Operation) bool {
			return reflect.DeepEqual(op, operation)
		}) {
			return false
		}
	}
	return true
}
