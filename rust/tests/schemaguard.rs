use schema_guard::{diff, has_breaking, Change, ChangeKind, Field, Schema};

fn find<'a>(changes: &'a [Change], field: &str) -> Option<&'a Change> {
    changes.iter().find(|c| c.field == field)
}

#[test]
fn no_changes() {
    let s = Schema::of([Field::new("id", "string"), Field::new("age", "int")]);
    let changes = diff(&s, &s);
    assert!(changes.is_empty());
    assert!(!has_breaking(&changes));
}

#[test]
fn removing_field_is_breaking() {
    let old = Schema::of([Field::new("id", "string"), Field::new("age", "int")]);
    let new = Schema::of([Field::new("id", "string")]);
    let changes = diff(&old, &new);
    let age = find(&changes, "age").expect("age change");
    assert_eq!(age.kind, ChangeKind::Breaking);
    assert_eq!(age.detail, "field removed");
    assert!(has_breaking(&changes));
}

#[test]
fn adding_optional_field_is_safe() {
    let old = Schema::of([Field::new("id", "string")]);
    let new = Schema::of([
        Field::new("id", "string"),
        Field::optional("nickname", "string"),
    ]);
    let changes = diff(&old, &new);
    let nick = find(&changes, "nickname").expect("nickname change");
    assert_eq!(nick.kind, ChangeKind::Safe);
    assert_eq!(nick.detail, "added optional field");
    assert!(!has_breaking(&changes));
}

#[test]
fn adding_required_field_is_breaking() {
    let old = Schema::of([Field::new("id", "string")]);
    let new = Schema::of([Field::new("id", "string"), Field::new("email", "string")]);
    let changes = diff(&old, &new);
    let email = find(&changes, "email").expect("email change");
    assert_eq!(email.kind, ChangeKind::Breaking);
    assert_eq!(email.detail, "added required field");
}

#[test]
fn field_new_defaults_to_required() {
    assert!(Field::new("id", "string").required);
    assert!(!Field::optional("id", "string").required);
}

#[test]
fn type_change_is_breaking() {
    let old = Schema::of([Field::new("age", "int")]);
    let new = Schema::of([Field::new("age", "string")]);
    let changes = diff(&old, &new);
    let age = find(&changes, "age").expect("age change");
    assert_eq!(age.kind, ChangeKind::Breaking);
    assert_eq!(age.detail, "type changed int -> string");
}

#[test]
fn required_to_optional_is_safe() {
    let old = Schema::of([Field::new("mid", "string")]);
    let new = Schema::of([Field::optional("mid", "string")]);
    let changes = diff(&old, &new);
    assert_eq!(changes.len(), 1);
    assert_eq!(changes[0].kind, ChangeKind::Safe);
    assert_eq!(changes[0].detail, "required -> optional");
    assert!(!has_breaking(&changes));
}

#[test]
fn optional_to_required_is_breaking() {
    let old = Schema::of([Field::optional("mid", "string")]);
    let new = Schema::of([Field::new("mid", "string")]);
    let changes = diff(&old, &new);
    assert_eq!(changes.len(), 1);
    assert_eq!(changes[0].kind, ChangeKind::Breaking);
    assert_eq!(changes[0].detail, "optional -> required");
}

#[test]
fn breaking_sorted_first() {
    let old = Schema::of([Field::new("keep", "string"), Field::new("drop", "int")]);
    let new = Schema::of([
        Field::optional("keep", "string"),
        Field::new("added", "string"),
    ]);
    let changes = diff(&old, &new);
    assert_eq!(changes[0].kind, ChangeKind::Breaking);
}

#[test]
fn breaking_then_safe_ordered_by_field() {
    let old = Schema::of([
        Field::new("keep", "string"),
        Field::new("drop", "int"),
        Field::new("zeta", "int"),
    ]);
    let new = Schema::of([
        Field::optional("keep", "string"),
        Field::new("added", "string"),
        Field::optional("beta", "string"),
        Field::new("zeta", "int"),
    ]);
    let changes = diff(&old, &new);
    let fields: Vec<&str> = changes.iter().map(|c| c.field.as_str()).collect();
    let kinds: Vec<ChangeKind> = changes.iter().map(|c| c.kind).collect();
    assert_eq!(fields, vec!["added", "drop", "beta", "keep"]);
    assert_eq!(
        kinds,
        vec![
            ChangeKind::Breaking,
            ChangeKind::Breaking,
            ChangeKind::Safe,
            ChangeKind::Safe
        ]
    );
}

#[test]
fn type_change_and_optional_to_required_yields_two_breaking() {
    let old = Schema::of([Field::optional("x", "int")]);
    let new = Schema::of([Field::new("x", "string")]);
    let changes = diff(&old, &new);
    assert_eq!(changes.len(), 2);
    assert_eq!(changes[0].detail, "type changed int -> string");
    assert_eq!(changes[1].detail, "optional -> required");
    assert_eq!(changes[0].kind, ChangeKind::Breaking);
    assert_eq!(changes[1].kind, ChangeKind::Breaking);
}

#[test]
fn type_change_and_required_to_optional_yields_breaking_and_safe() {
    let old = Schema::of([Field::new("y", "int")]);
    let new = Schema::of([Field::optional("y", "string")]);
    let changes = diff(&old, &new);
    assert_eq!(changes.len(), 2);
    // Breaking sorts before safe even though the safe change was detected first.
    assert_eq!(changes[0].kind, ChangeKind::Breaking);
    assert_eq!(changes[0].detail, "type changed int -> string");
    assert_eq!(changes[1].kind, ChangeKind::Safe);
    assert_eq!(changes[1].detail, "required -> optional");
}

#[test]
fn has_breaking_empty_is_false() {
    assert!(!has_breaking(&[]));
}

#[test]
fn multiple_removals_all_breaking() {
    let old = Schema::of([
        Field::new("a", "int"),
        Field::new("b", "int"),
        Field::new("c", "int"),
    ]);
    let new = Schema::of([Field::new("a", "int")]);
    let changes = diff(&old, &new);
    assert_eq!(changes.len(), 2);
    assert!(changes.iter().all(|c| c.kind == ChangeKind::Breaking));
    assert_eq!(changes[0].field, "b");
    assert_eq!(changes[1].field, "c");
}
