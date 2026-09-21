/**
 * SchemaGuard: API schema diff + backward-compatibility classifier.
 *
 * Diff two versions of an API/data schema and classify every change as SAFE or
 * BREAKING using the standard rules of schema evolution. Wire it into CI to
 * fail a PR that introduces a breaking change unless it is explicitly
 * acknowledged.
 *
 * Backward-compatibility rules (consumer-facing):
 *   - remove a field            -> BREAKING
 *   - add a required field      -> BREAKING
 *   - add an optional field     -> SAFE
 *   - change a field's type     -> BREAKING
 *   - required -> optional      -> SAFE
 *   - optional -> required      -> BREAKING
 */

/** Classifies whether a schema change breaks consumers. */
export enum ChangeKind {
  Safe = "safe",
  Breaking = "breaking",
}

/** A single named field in a schema. `required` defaults to true. */
export class Field {
  readonly name: string;
  readonly type: string;
  readonly required: boolean;

  constructor(name: string, type: string, required = true) {
    this.name = name;
    this.type = type;
    this.required = required;
  }
}

/** A set of fields keyed by name (insertion order preserved). */
export class Schema {
  readonly fields: Map<string, Field>;

  constructor(fields: Map<string, Field>) {
    this.fields = fields;
  }

  /** Builds a schema from the given fields, keyed by field name. */
  static of(...fields: Field[]): Schema {
    const map = new Map<string, Field>();
    for (const f of fields) {
      map.set(f.name, f);
    }
    return new Schema(map);
  }
}

/** A single classified difference between two schemas. */
export interface Change {
  readonly field: string;
  readonly kind: ChangeKind;
  readonly detail: string;
}

/**
 * Compares `old` and `next` schemas and returns every change, breaking changes
 * first and then ordered by field name.
 *
 * A single field may yield more than one change (for example a type change plus
 * optional -> required); those keep their detection order because the sort is
 * stable.
 */
export function diff(old: Schema, next: Schema): Change[] {
  const changes: Change[] = [];

  for (const [name, oldField] of old.fields) {
    const newField = next.fields.get(name);
    if (newField === undefined) {
      changes.push({ field: name, kind: ChangeKind.Breaking, detail: "field removed" });
      continue;
    }
    if (newField.type !== oldField.type) {
      changes.push({
        field: name,
        kind: ChangeKind.Breaking,
        detail: `type changed ${oldField.type} -> ${newField.type}`,
      });
    }
    if (!oldField.required && newField.required) {
      changes.push({ field: name, kind: ChangeKind.Breaking, detail: "optional -> required" });
    } else if (oldField.required && !newField.required) {
      changes.push({ field: name, kind: ChangeKind.Safe, detail: "required -> optional" });
    }
  }

  for (const [name, newField] of next.fields) {
    if (!old.fields.has(name)) {
      changes.push({
        field: name,
        kind: newField.required ? ChangeKind.Breaking : ChangeKind.Safe,
        detail: newField.required ? "added required field" : "added optional field",
      });
    }
  }

  changes.sort((a, b) => {
    const byKind = kindRank(a.kind) - kindRank(b.kind);
    if (byKind !== 0) {
      return byKind;
    }
    return a.field < b.field ? -1 : a.field > b.field ? 1 : 0;
  });
  return changes;
}

/** Reports whether any change is breaking. */
export function hasBreaking(changes: Change[]): boolean {
  return changes.some((c) => c.kind === ChangeKind.Breaking);
}

function kindRank(kind: ChangeKind): number {
  return kind === ChangeKind.Breaking ? 0 : 1;
}
