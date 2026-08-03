package com.schemaguard;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** Diffs two schemas and classifies each change as SAFE or BREAKING. */
public final class SchemaGuard {

    public enum ChangeKind {
        SAFE, BREAKING
    }

    public record Field(String name, String type, boolean required) {
        public Field(String name, String type) {
            this(name, type, true);
        }
    }

    public record Change(String field, ChangeKind kind, String detail) {
    }

    public static final class Schema {
        final Map<String, Field> fields = new LinkedHashMap<>();

        public static Schema of(Field... fields) {
            Schema s = new Schema();
            for (Field f : fields) {
                s.fields.put(f.name(), f);
            }
            return s;
        }
    }

    public static List<Change> diff(Schema old, Schema next) {
        List<Change> changes = new ArrayList<>();

        for (Map.Entry<String, Field> e : old.fields.entrySet()) {
            String name = e.getKey();
            Field of = e.getValue();
            Field nf = next.fields.get(name);
            if (nf == null) {
                changes.add(new Change(name, ChangeKind.BREAKING, "field removed"));
                continue;
            }
            if (!nf.type().equals(of.type())) {
                changes.add(new Change(name, ChangeKind.BREAKING,
                        "type changed " + of.type() + " -> " + nf.type()));
            }
            if (!of.required() && nf.required()) {
                changes.add(new Change(name, ChangeKind.BREAKING, "optional -> required"));
            } else if (of.required() && !nf.required()) {
                changes.add(new Change(name, ChangeKind.SAFE, "required -> optional"));
            }
        }

        for (Map.Entry<String, Field> e : next.fields.entrySet()) {
            if (!old.fields.containsKey(e.getKey())) {
                Field nf = e.getValue();
                ChangeKind kind = nf.required() ? ChangeKind.BREAKING : ChangeKind.SAFE;
                String detail = nf.required() ? "added required field" : "added optional field";
                changes.add(new Change(e.getKey(), kind, detail));
            }
        }

        changes.sort(Comparator
                .comparingInt((Change c) -> c.kind() == ChangeKind.BREAKING ? 0 : 1)
                .thenComparing(Change::field));
        return changes;
    }

    public static boolean hasBreaking(List<Change> changes) {
        return changes.stream().anyMatch(c -> c.kind() == ChangeKind.BREAKING);
    }

    private SchemaGuard() {
    }
}
