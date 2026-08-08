# schema-guard: design, trade-offs, and non-goals

Status: accepted
Author: Parag Sawant

Why schema-guard is built the way it is. It answers one question - "does this change
to an API or data schema break the people consuming it?" - and the whole value is in
getting the backward-compatibility rules right and catching the break at the pull
request instead of in production.

## Problem and goals

Every schema change looks harmless in the diff until a downstream consumer breaks on
it. schema-guard diffs two versions of a schema, classifies each change as SAFE or
BREAKING by the standard rules of schema evolution, and - wired into CI - fails a PR
that introduces a breaking change unless it's explicitly acknowledged. Goals:

1. Encode the **backward-compatibility rules** correctly and legibly, so a
   classification is something you can point at, not a black box.
2. Catch breaks **at the pull request**, as a CI gate, not after a deploy.
3. The **same rules in three languages** (Python, C#, Java), because the logic is the
   logic - a field removal is breaking regardless of which language checks it.

## The rules, stated plainly

Backward compatibility here means "an existing consumer keeps working." From that one
principle the rules fall out:

- **Remove a field** -> BREAKING (a consumer may depend on it).
- **Add a required field** -> BREAKING (existing producers won't send it).
- **Add an optional field** -> SAFE (old consumers ignore it).
- **Change a field's type** -> BREAKING (the contract shifts under the consumer).
- **Required -> optional** -> SAFE (a consumer that sent it still can).
- **Optional -> required** -> BREAKING (a consumer that omitted it now fails).

The asymmetry is the important part and the part that's easy to get wrong: loosening a
constraint (required -> optional) is safe, tightening one (optional -> required) is
not. Pinning that direction down is most of the point of the tool.

## Key design decisions

**Consumer-facing compatibility is the fixed reference frame.** "Breaking" always
means "breaks an existing consumer," never producer-side convenience. Choosing one
perspective and holding it is what makes the classification unambiguous - the same
change is always classified the same way, because the question is always the same.

**Pure diff, no I/O in the core.** `diff(old, new)` takes two schemas and returns a
list of classified changes; loading schemas from files is a separate CLI concern. That
keeps the rules trivially testable and makes the three-language ports line up exactly.

**The gate is `has_breaking` -> exit code.** The CLI turns "any breaking change" into a
non-zero exit, so it drops into any CI pipeline as a step that can fail the build. An
acknowledged break is handled by the workflow around it (an override), not by softening
the classification.

## Trade-offs I made on purpose

- **Structural rules, not semantic ones.** schema-guard reasons about fields, types,
  and required-ness. It does not understand that renaming `qty` to `quantity` while
  changing nothing else is "really" one rename - it sees a remove plus an add, and
  flags the remove as breaking. That's the safe reading (a consumer of `qty` does
  break), and treating a rename specially would require intent the diff doesn't have.
  Noted as the deliberate conservative choice.
- **Type equality is exact.** A type change is breaking, full stop; the tool doesn't
  model "widening" conversions (int -> long) as safe, because whether that's actually
  safe depends on the wire format and the consumer. Exact-match keeps it honest and
  predictable, and errs toward flagging.
- **Flat field model.** The core compares fields by name. Deeply nested or
  polymorphic schemas would need a richer model; the flat version covers the common
  case and keeps the rules readable. A structural extension is the natural next step.

## Why there's no benchmark here

Diffing two schemas is a handful of dictionary comparisons that runs once per CI check
in microseconds - throughput is not a property anyone cares about, and reporting one
would imply a performance concern that doesn't exist. Correctness of the
classification is the entire game, and that's what the tri-language test suites pin
down: every rule above, in all three languages, asserted identically.

## Non-goals

- **Not a schema registry.** It compares two schemas you hand it; it doesn't store,
  version, or serve them.
- **Not a migration generator.** It tells you a change is breaking; it doesn't write
  the migration or the compatibility shim.
- **Not a semantic differ.** No rename detection or intent inference - by design, it
  reads changes conservatively from the consumer's side.

Part of [parag-labs](https://github.com/parag-labs) - small, focused tools for building AI systems you can trust.
