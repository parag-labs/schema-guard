// Package schemaguard diffs two versions of an API/data schema and classifies
// every change as SAFE or BREAKING using the standard rules of schema
// evolution.
//
// Backward-compatibility rules (consumer-facing):
//
//	remove a field            -> BREAKING
//	add a required field      -> BREAKING
//	add an optional field     -> SAFE
//	change a field's type     -> BREAKING
//	required -> optional      -> SAFE
//	optional -> required      -> BREAKING
package schemaguard

import "sort"

// ChangeKind classifies whether a schema change breaks consumers.
type ChangeKind string

const (
	// Safe changes are backward compatible.
	Safe ChangeKind = "safe"
	// Breaking changes are not backward compatible.
	Breaking ChangeKind = "breaking"
)

// Field is a single named field in a schema. Required defaults to true when a
// field is built with NewField.
type Field struct {
	// Name is the field's identifier.
	Name string
	// Type is the field's declared type.
	Type string
	// Required reports whether consumers must supply the field.
	Required bool
}

// NewField builds a Field. Required defaults to true (matching the reference);
// pass an explicit bool to override, e.g. NewField("nickname", "string", false).
func NewField(name, typ string, required ...bool) Field {
	req := true
	if len(required) > 0 {
		req = required[0]
	}
	return Field{Name: name, Type: typ, Required: req}
}

// Schema is a set of fields keyed by name.
type Schema struct {
	// Fields maps each field name to its definition.
	Fields map[string]Field
}

// SchemaOf builds a Schema from the given fields, keyed by field name.
func SchemaOf(fields ...Field) Schema {
	m := make(map[string]Field, len(fields))
	for _, f := range fields {
		m[f.Name] = f
	}
	return Schema{Fields: m}
}

// Change is a single classified difference between two schemas.
type Change struct {
	// Field is the name of the field that changed.
	Field string
	// Kind is the change's backward-compatibility classification.
	Kind ChangeKind
	// Detail is a human-readable description of the change.
	Detail string
}

// Diff compares old and new schemas and returns every change, breaking changes
// first and then ordered by field name. A single field may yield more than one
// change (for example a type change plus optional -> required); those keep
// their detection order.
func Diff(old, new Schema) []Change {
	changes := []Change{}

	for name, of := range old.Fields {
		nf, ok := new.Fields[name]
		if !ok {
			changes = append(changes, Change{name, Breaking, "field removed"})
			continue
		}
		if nf.Type != of.Type {
			changes = append(changes, Change{name, Breaking, "type changed " + of.Type + " -> " + nf.Type})
		}
		if !of.Required && nf.Required {
			changes = append(changes, Change{name, Breaking, "optional -> required"})
		} else if of.Required && !nf.Required {
			changes = append(changes, Change{name, Safe, "required -> optional"})
		}
	}

	for name, nf := range new.Fields {
		if _, ok := old.Fields[name]; !ok {
			if nf.Required {
				changes = append(changes, Change{name, Breaking, "added required field"})
			} else {
				changes = append(changes, Change{name, Safe, "added optional field"})
			}
		}
	}

	sort.SliceStable(changes, func(i, j int) bool {
		ri, rj := kindRank(changes[i].Kind), kindRank(changes[j].Kind)
		if ri != rj {
			return ri < rj
		}
		return changes[i].Field < changes[j].Field
	})
	return changes
}

// HasBreaking reports whether any change is breaking.
func HasBreaking(changes []Change) bool {
	for _, c := range changes {
		if c.Kind == Breaking {
			return true
		}
	}
	return false
}

func kindRank(k ChangeKind) int {
	if k == Breaking {
		return 0
	}
	return 1
}
