# SchemaGuard

**Stop breaking API changes at the PR, not in production.**

A backend team renames a field or makes an optional field required, and three downstream services fall over. Code review didn't catch it because nobody diffs schemas by hand. SchemaGuard diffs two versions of a schema, classifies every change as **safe** or **breaking** using standard backward-compatibility rules, and **fails the PR** on a breaking change. Same rules in **Python, C#, and Java**.

## The rules (consumer backward compatibility)

| Change | Verdict |
|--------|:-------:|
| Remove a field | **breaking** |
| Add a required field | **breaking** |
| Add an optional field | safe |
| Change a field's type | **breaking** |
| Required → optional | safe |
| Optional → required | **breaking** |

## Use it (Python)

```bash
cd python
python -m src.cli old_schema.json new_schema.json
#   [BREAKING] email: added required field
#   [safe] nickname: added optional field
# FAIL: breaking schema changes detected   (exit 1)
```

Breaking changes sort first so they're impossible to miss in CI logs.

## Three languages, one ruleset

| Language | Tests | Run |
|----------|:-----:|-----|
| Python | 8 | `cd python && pytest -q` |
| C# (.NET 10) | 8 | `cd csharp && dotnet test` |
| Java (17+) | 8 | `cd java && mvn test` |

## Known limitations / next

- Works on a normalized field model — parsers for OpenAPI, protobuf, and SQL `ALTER TABLE` migrations feed into it next.
- Type compatibility is exact-match; a widening-vs-narrowing type lattice (e.g. `int32 -> int64` is safe) is a natural refinement.
- No "acknowledged breaking change" escape hatch yet — a `--allow-breaking` flag with a changelog entry would round it out.

Part of [parag-labs](https://github.com/parag-labs) — small, focused tools for building AI systems you can trust.
