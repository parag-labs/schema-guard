package schemaguard

import "testing"

func fieldByName(changes []Change, name string) (Change, bool) {
	for _, c := range changes {
		if c.Field == name {
			return c, true
		}
	}
	return Change{}, false
}

func TestNoChanges(t *testing.T) {
	s := SchemaOf(NewField("id", "string"), NewField("age", "int"))
	changes := Diff(s, s)
	if len(changes) != 0 {
		t.Fatalf("expected no changes, got %v", changes)
	}
	if HasBreaking(changes) {
		t.Fatal("expected no breaking changes")
	}
}

func TestRemovingFieldIsBreaking(t *testing.T) {
	old := SchemaOf(NewField("id", "string"), NewField("age", "int"))
	new := SchemaOf(NewField("id", "string"))
	changes := Diff(old, new)
	c, ok := fieldByName(changes, "age")
	if !ok || c.Kind != Breaking {
		t.Fatalf("expected age removal to be breaking, got %v", changes)
	}
	if c.Detail != "field removed" {
		t.Fatalf("detail = %q, want 'field removed'", c.Detail)
	}
	if !HasBreaking(changes) {
		t.Fatal("expected a breaking change")
	}
}

func TestAddingOptionalFieldIsSafe(t *testing.T) {
	old := SchemaOf(NewField("id", "string"))
	new := SchemaOf(NewField("id", "string"), NewField("nickname", "string", false))
	changes := Diff(old, new)
	c, ok := fieldByName(changes, "nickname")
	if !ok || c.Kind != Safe {
		t.Fatalf("expected nickname add to be safe, got %v", changes)
	}
	if c.Detail != "added optional field" {
		t.Fatalf("detail = %q, want 'added optional field'", c.Detail)
	}
	if HasBreaking(changes) {
		t.Fatal("expected no breaking change")
	}
}

func TestAddingRequiredFieldIsBreaking(t *testing.T) {
	old := SchemaOf(NewField("id", "string"))
	new := SchemaOf(NewField("id", "string"), NewField("email", "string", true))
	changes := Diff(old, new)
	c, ok := fieldByName(changes, "email")
	if !ok || c.Kind != Breaking {
		t.Fatalf("expected email add to be breaking, got %v", changes)
	}
	if c.Detail != "added required field" {
		t.Fatalf("detail = %q, want 'added required field'", c.Detail)
	}
}

func TestNewFieldDefaultsToRequired(t *testing.T) {
	f := NewField("id", "string")
	if !f.Required {
		t.Fatal("expected NewField to default Required to true")
	}
}

func TestTypeChangeIsBreaking(t *testing.T) {
	old := SchemaOf(NewField("age", "int"))
	new := SchemaOf(NewField("age", "string"))
	changes := Diff(old, new)
	c, ok := fieldByName(changes, "age")
	if !ok || c.Kind != Breaking {
		t.Fatalf("expected age type change to be breaking, got %v", changes)
	}
	if c.Detail != "type changed int -> string" {
		t.Fatalf("detail = %q, want 'type changed int -> string'", c.Detail)
	}
}

func TestRequiredToOptionalIsSafe(t *testing.T) {
	old := SchemaOf(NewField("mid", "string", true))
	new := SchemaOf(NewField("mid", "string", false))
	changes := Diff(old, new)
	if len(changes) != 1 || changes[0].Kind != Safe {
		t.Fatalf("expected one safe change, got %v", changes)
	}
	if changes[0].Detail != "required -> optional" {
		t.Fatalf("detail = %q, want 'required -> optional'", changes[0].Detail)
	}
	if HasBreaking(changes) {
		t.Fatal("expected no breaking change")
	}
}

func TestOptionalToRequiredIsBreaking(t *testing.T) {
	old := SchemaOf(NewField("mid", "string", false))
	new := SchemaOf(NewField("mid", "string", true))
	changes := Diff(old, new)
	if len(changes) != 1 || changes[0].Kind != Breaking {
		t.Fatalf("expected one breaking change, got %v", changes)
	}
	if changes[0].Detail != "optional -> required" {
		t.Fatalf("detail = %q, want 'optional -> required'", changes[0].Detail)
	}
}

func TestBreakingSortedFirst(t *testing.T) {
	old := SchemaOf(NewField("keep", "string", true), NewField("drop", "int"))
	new := SchemaOf(NewField("keep", "string", false), NewField("added", "string", true))
	changes := Diff(old, new)
	if changes[0].Kind != Breaking {
		t.Fatalf("expected a breaking change first, got %v", changes)
	}
}

func TestBreakingThenSafeOrderedByField(t *testing.T) {
	old := SchemaOf(
		NewField("keep", "string", true),
		NewField("drop", "int"),
		NewField("zeta", "int"),
	)
	new := SchemaOf(
		NewField("keep", "string", false),
		NewField("added", "string", true),
		NewField("beta", "string", false),
		NewField("zeta", "int"),
	)
	changes := Diff(old, new)
	kinds := make([]ChangeKind, len(changes))
	fields := make([]string, len(changes))
	for i, c := range changes {
		kinds[i] = c.Kind
		fields[i] = c.Field
	}
	wantFields := []string{"added", "drop", "beta", "keep"}
	wantKinds := []ChangeKind{Breaking, Breaking, Safe, Safe}
	for i := range wantFields {
		if fields[i] != wantFields[i] || kinds[i] != wantKinds[i] {
			t.Fatalf("order = %v/%v, want %v/%v", fields, kinds, wantFields, wantKinds)
		}
	}
}

func TestTypeChangeAndOptionalToRequiredYieldsTwoBreaking(t *testing.T) {
	old := SchemaOf(NewField("x", "int", false))
	new := SchemaOf(NewField("x", "string", true))
	changes := Diff(old, new)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %v", changes)
	}
	if changes[0].Detail != "type changed int -> string" {
		t.Fatalf("first change = %q, want type change first", changes[0].Detail)
	}
	if changes[1].Detail != "optional -> required" {
		t.Fatalf("second change = %q, want 'optional -> required'", changes[1].Detail)
	}
	if changes[0].Kind != Breaking || changes[1].Kind != Breaking {
		t.Fatalf("expected both breaking, got %v", changes)
	}
}

func TestTypeChangeAndRequiredToOptionalYieldsBreakingAndSafe(t *testing.T) {
	old := SchemaOf(NewField("y", "int", true))
	new := SchemaOf(NewField("y", "string", false))
	changes := Diff(old, new)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %v", changes)
	}
	// Breaking sorts before safe even though the safe change was detected first.
	if changes[0].Kind != Breaking || changes[0].Detail != "type changed int -> string" {
		t.Fatalf("first change = %v, want breaking type change", changes[0])
	}
	if changes[1].Kind != Safe || changes[1].Detail != "required -> optional" {
		t.Fatalf("second change = %v, want safe required->optional", changes[1])
	}
}

func TestHasBreakingEmpty(t *testing.T) {
	if HasBreaking([]Change{}) {
		t.Fatal("expected HasBreaking of empty to be false")
	}
}

func TestMultipleRemovalsAllBreaking(t *testing.T) {
	old := SchemaOf(NewField("a", "int"), NewField("b", "int"), NewField("c", "int"))
	new := SchemaOf(NewField("a", "int"))
	changes := Diff(old, new)
	if len(changes) != 2 {
		t.Fatalf("expected 2 removals, got %v", changes)
	}
	for _, c := range changes {
		if c.Kind != Breaking {
			t.Fatalf("expected all breaking, got %v", changes)
		}
	}
	if changes[0].Field != "b" || changes[1].Field != "c" {
		t.Fatalf("expected removals ordered b,c, got %v", changes)
	}
}
