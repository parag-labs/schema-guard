"""API schema diff + backward-compatibility classifier.

Diff two versions of an API/data schema and classify every change as SAFE or
BREAKING using the standard rules of schema evolution. Wire it into CI to fail a
PR that introduces a breaking change unless it's explicitly acknowledged.

Rules (consumer-facing backward compatibility):
  - remove a field            -> BREAKING
  - add a required field      -> BREAKING
  - add an optional field     -> SAFE
  - change a field's type     -> BREAKING
  - required -> optional      -> SAFE
  - optional -> required      -> BREAKING
"""

from __future__ import annotations

from dataclasses import dataclass, field as dc_field
from enum import Enum


class ChangeKind(str, Enum):
    SAFE = "safe"
    BREAKING = "breaking"


@dataclass(frozen=True)
class Field:
    name: str
    type: str
    required: bool = True


@dataclass
class Schema:
    fields: dict[str, Field] = dc_field(default_factory=dict)

    @classmethod
    def of(cls, *fields: Field) -> "Schema":
        return cls({f.name: f for f in fields})


@dataclass(frozen=True)
class Change:
    field: str
    kind: ChangeKind
    detail: str


def diff(old: Schema, new: Schema) -> list[Change]:
    changes: list[Change] = []

    for name, of in old.fields.items():
        nf = new.fields.get(name)
        if nf is None:
            changes.append(Change(name, ChangeKind.BREAKING, "field removed"))
            continue
        if nf.type != of.type:
            changes.append(
                Change(name, ChangeKind.BREAKING, f"type changed {of.type} -> {nf.type}")
            )
        if not of.required and nf.required:
            changes.append(Change(name, ChangeKind.BREAKING, "optional -> required"))
        elif of.required and not nf.required:
            changes.append(Change(name, ChangeKind.SAFE, "required -> optional"))

    for name, nf in new.fields.items():
        if name not in old.fields:
            kind = ChangeKind.BREAKING if nf.required else ChangeKind.SAFE
            detail = "added required field" if nf.required else "added optional field"
            changes.append(Change(name, kind, detail))

    return sorted(changes, key=lambda c: (c.kind is not ChangeKind.BREAKING, c.field))


def has_breaking(changes: list[Change]) -> bool:
    return any(c.kind is ChangeKind.BREAKING for c in changes)
