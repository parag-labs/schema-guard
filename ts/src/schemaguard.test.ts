import { describe, expect, it } from "vitest";
import {
  Change,
  ChangeKind,
  Field,
  Schema,
  diff,
  hasBreaking,
} from "./schemaguard";

function findChange(changes: Change[], name: string): Change | undefined {
  return changes.find((c) => c.field === name);
}

describe("diff", () => {
  it("reports no changes for identical schemas", () => {
    const s = Schema.of(new Field("id", "string"), new Field("age", "int"));
    const changes = diff(s, s);
    expect(changes).toHaveLength(0);
    expect(hasBreaking(changes)).toBe(false);
  });

  it("treats removing a field as breaking", () => {
    const before = Schema.of(new Field("id", "string"), new Field("age", "int"));
    const after = Schema.of(new Field("id", "string"));
    const changes = diff(before, after);
    const c = findChange(changes, "age");
    expect(c?.kind).toBe(ChangeKind.Breaking);
    expect(c?.detail).toBe("field removed");
    expect(hasBreaking(changes)).toBe(true);
  });

  it("treats adding an optional field as safe", () => {
    const before = Schema.of(new Field("id", "string"));
    const after = Schema.of(new Field("id", "string"), new Field("nickname", "string", false));
    const changes = diff(before, after);
    const c = findChange(changes, "nickname");
    expect(c?.kind).toBe(ChangeKind.Safe);
    expect(c?.detail).toBe("added optional field");
    expect(hasBreaking(changes)).toBe(false);
  });

  it("treats adding a required field as breaking", () => {
    const before = Schema.of(new Field("id", "string"));
    const after = Schema.of(new Field("id", "string"), new Field("email", "string", true));
    const changes = diff(before, after);
    const c = findChange(changes, "email");
    expect(c?.kind).toBe(ChangeKind.Breaking);
    expect(c?.detail).toBe("added required field");
  });

  it("defaults Field.required to true", () => {
    const f = new Field("id", "string");
    expect(f.required).toBe(true);
  });

  it("treats a type change as breaking", () => {
    const before = Schema.of(new Field("age", "int"));
    const after = Schema.of(new Field("age", "string"));
    const changes = diff(before, after);
    const c = findChange(changes, "age");
    expect(c?.kind).toBe(ChangeKind.Breaking);
    expect(c?.detail).toBe("type changed int -> string");
  });

  it("treats required -> optional as safe", () => {
    const before = Schema.of(new Field("mid", "string", true));
    const after = Schema.of(new Field("mid", "string", false));
    const changes = diff(before, after);
    expect(changes).toHaveLength(1);
    expect(changes[0].kind).toBe(ChangeKind.Safe);
    expect(changes[0].detail).toBe("required -> optional");
    expect(hasBreaking(changes)).toBe(false);
  });

  it("treats optional -> required as breaking", () => {
    const before = Schema.of(new Field("mid", "string", false));
    const after = Schema.of(new Field("mid", "string", true));
    const changes = diff(before, after);
    expect(changes).toHaveLength(1);
    expect(changes[0].kind).toBe(ChangeKind.Breaking);
    expect(changes[0].detail).toBe("optional -> required");
  });

  it("sorts breaking changes first", () => {
    const before = Schema.of(new Field("keep", "string", true), new Field("drop", "int"));
    const after = Schema.of(new Field("keep", "string", false), new Field("added", "string", true));
    const changes = diff(before, after);
    expect(changes[0].kind).toBe(ChangeKind.Breaking);
  });

  it("orders breaking group then safe group, each by field name", () => {
    const before = Schema.of(
      new Field("keep", "string", true),
      new Field("drop", "int"),
      new Field("zeta", "int"),
    );
    const after = Schema.of(
      new Field("keep", "string", false),
      new Field("added", "string", true),
      new Field("beta", "string", false),
      new Field("zeta", "int"),
    );
    const changes = diff(before, after);
    expect(changes.map((c) => c.field)).toEqual(["added", "drop", "beta", "keep"]);
    expect(changes.map((c) => c.kind)).toEqual([
      ChangeKind.Breaking,
      ChangeKind.Breaking,
      ChangeKind.Safe,
      ChangeKind.Safe,
    ]);
  });

  it("yields two breaking changes for type change plus optional -> required, type first", () => {
    const before = Schema.of(new Field("x", "int", false));
    const after = Schema.of(new Field("x", "string", true));
    const changes = diff(before, after);
    expect(changes).toHaveLength(2);
    expect(changes[0].detail).toBe("type changed int -> string");
    expect(changes[1].detail).toBe("optional -> required");
    expect(changes[0].kind).toBe(ChangeKind.Breaking);
    expect(changes[1].kind).toBe(ChangeKind.Breaking);
  });

  it("keeps breaking type change before safe required -> optional despite detection order", () => {
    const before = Schema.of(new Field("y", "int", true));
    const after = Schema.of(new Field("y", "string", false));
    const changes = diff(before, after);
    expect(changes).toHaveLength(2);
    expect(changes[0].kind).toBe(ChangeKind.Breaking);
    expect(changes[0].detail).toBe("type changed int -> string");
    expect(changes[1].kind).toBe(ChangeKind.Safe);
    expect(changes[1].detail).toBe("required -> optional");
  });
});

describe("hasBreaking", () => {
  it("is false for an empty change list", () => {
    expect(hasBreaking([])).toBe(false);
  });

  it("treats multiple removals as all breaking, ordered by field name", () => {
    const before = Schema.of(new Field("a", "int"), new Field("b", "int"), new Field("c", "int"));
    const after = Schema.of(new Field("a", "int"));
    const changes = diff(before, after);
    expect(changes).toHaveLength(2);
    expect(changes.every((c) => c.kind === ChangeKind.Breaking)).toBe(true);
    expect(changes.map((c) => c.field)).toEqual(["b", "c"]);
  });
});
