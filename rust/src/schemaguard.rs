//! Schema diffing and backward-compatibility classification.

use std::collections::HashMap;

/// Classifies whether a schema change breaks consumers.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ChangeKind {
    /// Backward-compatible change.
    Safe,
    /// Not backward-compatible change.
    Breaking,
}

/// A single named field in a schema.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Field {
    /// The field's identifier.
    pub name: String,
    /// The field's declared type.
    pub r#type: String,
    /// Whether consumers must supply the field.
    pub required: bool,
}

impl Field {
    /// Builds a required field (the reference default).
    pub fn new(name: impl Into<String>, ty: impl Into<String>) -> Self {
        Field {
            name: name.into(),
            r#type: ty.into(),
            required: true,
        }
    }

    /// Builds an optional field.
    pub fn optional(name: impl Into<String>, ty: impl Into<String>) -> Self {
        Field {
            name: name.into(),
            r#type: ty.into(),
            required: false,
        }
    }
}

/// A set of fields keyed by name.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct Schema {
    /// Maps each field name to its definition.
    pub fields: HashMap<String, Field>,
}

impl Schema {
    /// Builds a schema from the given fields, keyed by field name.
    pub fn of(fields: impl IntoIterator<Item = Field>) -> Self {
        let mut map = HashMap::new();
        for f in fields {
            map.insert(f.name.clone(), f);
        }
        Schema { fields: map }
    }
}

/// A single classified difference between two schemas.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Change {
    /// Name of the field that changed.
    pub field: String,
    /// Backward-compatibility classification.
    pub kind: ChangeKind,
    /// Human-readable description of the change.
    pub detail: String,
}

/// Compares `old` and `new` schemas and returns every change, breaking changes
/// first and then ordered by field name.
///
/// A single field may yield more than one change (for example a type change
/// plus optional -> required); those keep their detection order because the
/// sort is stable.
#[must_use]
pub fn diff(old: &Schema, new: &Schema) -> Vec<Change> {
    let mut changes: Vec<Change> = Vec::new();

    for (name, of) in &old.fields {
        let Some(nf) = new.fields.get(name) else {
            changes.push(Change {
                field: name.clone(),
                kind: ChangeKind::Breaking,
                detail: "field removed".to_string(),
            });
            continue;
        };
        if nf.r#type != of.r#type {
            changes.push(Change {
                field: name.clone(),
                kind: ChangeKind::Breaking,
                detail: format!("type changed {} -> {}", of.r#type, nf.r#type),
            });
        }
        if !of.required && nf.required {
            changes.push(Change {
                field: name.clone(),
                kind: ChangeKind::Breaking,
                detail: "optional -> required".to_string(),
            });
        } else if of.required && !nf.required {
            changes.push(Change {
                field: name.clone(),
                kind: ChangeKind::Safe,
                detail: "required -> optional".to_string(),
            });
        }
    }

    for (name, nf) in &new.fields {
        if !old.fields.contains_key(name) {
            let (kind, detail) = if nf.required {
                (ChangeKind::Breaking, "added required field")
            } else {
                (ChangeKind::Safe, "added optional field")
            };
            changes.push(Change {
                field: name.clone(),
                kind,
                detail: detail.to_string(),
            });
        }
    }

    changes.sort_by(|a, b| {
        kind_rank(a.kind)
            .cmp(&kind_rank(b.kind))
            .then_with(|| a.field.cmp(&b.field))
    });
    changes
}

/// Reports whether any change is breaking.
#[must_use]
pub fn has_breaking(changes: &[Change]) -> bool {
    changes.iter().any(|c| c.kind == ChangeKind::Breaking)
}

fn kind_rank(kind: ChangeKind) -> u8 {
    match kind {
        ChangeKind::Breaking => 0,
        ChangeKind::Safe => 1,
    }
}
