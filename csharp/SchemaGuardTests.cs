using Xunit;

namespace SchemaGuard.Tests;

public class SchemaGuardTests
{
    [Fact]
    public void NoChanges()
    {
        var s = Schema.Of(new Field("id", "string"), new Field("age", "int"));
        Assert.Empty(SchemaDiff.Diff(s, s));
        Assert.False(SchemaDiff.HasBreaking(SchemaDiff.Diff(s, s)));
    }

    [Fact]
    public void RemovingFieldIsBreaking()
    {
        var old = Schema.Of(new Field("id", "string"), new Field("age", "int"));
        var @new = Schema.Of(new Field("id", "string"));
        var changes = SchemaDiff.Diff(old, @new);
        Assert.Contains(changes, c => c.Field == "age" && c.Kind == ChangeKind.Breaking);
        Assert.True(SchemaDiff.HasBreaking(changes));
    }

    [Fact]
    public void AddingOptionalFieldIsSafe()
    {
        var old = Schema.Of(new Field("id", "string"));
        var @new = Schema.Of(new Field("id", "string"), new Field("nickname", "string", false));
        var changes = SchemaDiff.Diff(old, @new);
        Assert.Contains(changes, c => c.Field == "nickname" && c.Kind == ChangeKind.Safe);
        Assert.False(SchemaDiff.HasBreaking(changes));
    }

    [Fact]
    public void AddingRequiredFieldIsBreaking()
    {
        var old = Schema.Of(new Field("id", "string"));
        var @new = Schema.Of(new Field("id", "string"), new Field("email", "string", true));
        var changes = SchemaDiff.Diff(old, @new);
        Assert.Contains(changes, c => c.Field == "email" && c.Kind == ChangeKind.Breaking);
    }

    [Fact]
    public void TypeChangeIsBreaking()
    {
        var old = Schema.Of(new Field("age", "int"));
        var @new = Schema.Of(new Field("age", "string"));
        var changes = SchemaDiff.Diff(old, @new);
        Assert.Contains(changes, c => c.Detail.Contains("type changed") && c.Kind == ChangeKind.Breaking);
    }

    [Fact]
    public void RequiredToOptionalIsSafe()
    {
        var old = Schema.Of(new Field("mid", "string", true));
        var @new = Schema.Of(new Field("mid", "string", false));
        var changes = SchemaDiff.Diff(old, @new);
        Assert.Equal(ChangeKind.Safe, changes[0].Kind);
        Assert.False(SchemaDiff.HasBreaking(changes));
    }

    [Fact]
    public void OptionalToRequiredIsBreaking()
    {
        var old = Schema.Of(new Field("mid", "string", false));
        var @new = Schema.Of(new Field("mid", "string", true));
        var changes = SchemaDiff.Diff(old, @new);
        Assert.Contains(changes, c => c.Kind == ChangeKind.Breaking);
    }

    [Fact]
    public void BreakingSortedFirst()
    {
        var old = Schema.Of(new Field("keep", "string", true), new Field("drop", "int"));
        var @new = Schema.Of(new Field("keep", "string", false), new Field("added", "string", true));
        var changes = SchemaDiff.Diff(old, @new);
        Assert.Equal(ChangeKind.Breaking, changes[0].Kind);
    }
}
