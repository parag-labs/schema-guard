"""SchemaGuard tests: every backward-compatibility rule + the breaking gate."""

from schemaguard import ChangeKind, Field, Schema, diff, has_breaking


def test_no_changes():
    s = Schema.of(Field("id", "string"), Field("age", "int"))
    assert diff(s, s) == []
    assert has_breaking(diff(s, s)) is False


def test_removing_field_is_breaking():
    old = Schema.of(Field("id", "string"), Field("age", "int"))
    new = Schema.of(Field("id", "string"))
    changes = diff(old, new)
    assert any(c.field == "age" and c.kind is ChangeKind.BREAKING for c in changes)
    assert has_breaking(changes)


def test_adding_optional_field_is_safe():
    old = Schema.of(Field("id", "string"))
    new = Schema.of(Field("id", "string"), Field("nickname", "string", required=False))
    changes = diff(old, new)
    assert any(c.field == "nickname" and c.kind is ChangeKind.SAFE for c in changes)
    assert has_breaking(changes) is False


def test_adding_required_field_is_breaking():
    old = Schema.of(Field("id", "string"))
    new = Schema.of(Field("id", "string"), Field("email", "string", required=True))
    changes = diff(old, new)
    assert any(c.field == "email" and c.kind is ChangeKind.BREAKING for c in changes)


def test_type_change_is_breaking():
    old = Schema.of(Field("age", "int"))
    new = Schema.of(Field("age", "string"))
    changes = diff(old, new)
    assert any("type changed" in c.detail and c.kind is ChangeKind.BREAKING for c in changes)


def test_required_to_optional_is_safe():
    old = Schema.of(Field("mid", "string", required=True))
    new = Schema.of(Field("mid", "string", required=False))
    changes = diff(old, new)
    assert changes[0].kind is ChangeKind.SAFE
    assert has_breaking(changes) is False


def test_optional_to_required_is_breaking():
    old = Schema.of(Field("mid", "string", required=False))
    new = Schema.of(Field("mid", "string", required=True))
    changes = diff(old, new)
    assert any(c.kind is ChangeKind.BREAKING for c in changes)


def test_breaking_sorted_first():
    old = Schema.of(Field("keep", "string", required=True), Field("drop", "int"))
    new = Schema.of(
        Field("keep", "string", required=False),      # safe
        Field("added", "string", required=True),       # breaking
    )
    changes = diff(old, new)
    # Breaking changes should come first for visibility.
    assert changes[0].kind is ChangeKind.BREAKING
