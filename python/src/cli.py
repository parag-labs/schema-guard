"""SchemaGuard CLI: fail the build when a schema change is backward-incompatible.

    python -m cli old.json new.json

Each JSON file is {"fields": [{"name","type","required"}, ...]}.
"""

from __future__ import annotations

import json
import sys

from schemaguard import Field, Schema, diff, has_breaking


def load(path: str) -> Schema:
    with open(path, encoding="utf-8") as f:
        data = json.load(f)
    return Schema.of(*[
        Field(x["name"], x["type"], bool(x.get("required", True))) for x in data["fields"]
    ])


def main(argv: list[str] | None = None) -> int:
    argv = argv if argv is not None else sys.argv[1:]
    if len(argv) != 2:
        print("usage: python -m cli <old.json> <new.json>")
        return 2

    changes = diff(load(argv[0]), load(argv[1]))
    if not changes:
        print("no schema changes")
        return 0

    for c in changes:
        mark = "BREAKING" if c.kind.value == "breaking" else "safe"
        print(f"  [{mark}] {c.field}: {c.detail}")

    if has_breaking(changes):
        print("\nFAIL: breaking schema changes detected")
        return 1
    print("\nOK: all changes are backward-compatible")
    return 0


if __name__ == "__main__":
    sys.exit(main())
