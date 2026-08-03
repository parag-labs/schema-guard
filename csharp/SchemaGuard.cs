namespace SchemaGuard;

public enum ChangeKind
{
    Safe,
    Breaking,
}

public readonly record struct Field(string Name, string Type, bool Required = true);

public sealed record Change(string Field, ChangeKind Kind, string Detail);

public sealed class Schema
{
    public Dictionary<string, Field> Fields { get; }

    public Schema(IEnumerable<Field> fields)
    {
        Fields = fields.ToDictionary(f => f.Name, f => f);
    }

    public static Schema Of(params Field[] fields) => new(fields);
}

/// <summary>Diffs two schemas and classifies each change as Safe or Breaking.</summary>
public static class SchemaDiff
{
    public static List<Change> Diff(Schema old, Schema @new)
    {
        var changes = new List<Change>();

        foreach (var (name, of) in old.Fields)
        {
            if (!@new.Fields.TryGetValue(name, out var nf))
            {
                changes.Add(new Change(name, ChangeKind.Breaking, "field removed"));
                continue;
            }

            if (nf.Type != of.Type)
            {
                changes.Add(new Change(name, ChangeKind.Breaking, $"type changed {of.Type} -> {nf.Type}"));
            }

            if (!of.Required && nf.Required)
            {
                changes.Add(new Change(name, ChangeKind.Breaking, "optional -> required"));
            }
            else if (of.Required && !nf.Required)
            {
                changes.Add(new Change(name, ChangeKind.Safe, "required -> optional"));
            }
        }

        foreach (var (name, nf) in @new.Fields)
        {
            if (!old.Fields.ContainsKey(name))
            {
                var kind = nf.Required ? ChangeKind.Breaking : ChangeKind.Safe;
                var detail = nf.Required ? "added required field" : "added optional field";
                changes.Add(new Change(name, kind, detail));
            }
        }

        return changes
            .OrderBy(c => c.Kind == ChangeKind.Breaking ? 0 : 1)
            .ThenBy(c => c.Field)
            .ToList();
    }

    public static bool HasBreaking(IEnumerable<Change> changes)
        => changes.Any(c => c.Kind == ChangeKind.Breaking);
}
