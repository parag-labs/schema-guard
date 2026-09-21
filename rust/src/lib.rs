//! SchemaGuard: API schema diff + backward-compatibility classifier.
//!
//! Diff two versions of an API/data schema and classify every change as SAFE or
//! BREAKING using the standard rules of schema evolution. Wire it into CI to
//! fail a PR that introduces a breaking change unless it is explicitly
//! acknowledged.
//!
//! Backward-compatibility rules (consumer-facing):
//!
//! - remove a field            -> BREAKING
//! - add a required field      -> BREAKING
//! - add an optional field     -> SAFE
//! - change a field's type     -> BREAKING
//! - required -> optional      -> SAFE
//! - optional -> required      -> BREAKING

pub mod schemaguard;

pub use schemaguard::{diff, has_breaking, Change, ChangeKind, Field, Schema};
