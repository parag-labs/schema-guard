package com.schemaguard;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.List;

import org.junit.jupiter.api.Test;

import com.schemaguard.SchemaGuard.Change;
import com.schemaguard.SchemaGuard.ChangeKind;
import com.schemaguard.SchemaGuard.Field;
import com.schemaguard.SchemaGuard.Schema;

class SchemaGuardTest {

    @Test
    void noChanges() {
        Schema s = Schema.of(new Field("id", "string"), new Field("age", "int"));
        assertTrue(SchemaGuard.diff(s, s).isEmpty());
        assertFalse(SchemaGuard.hasBreaking(SchemaGuard.diff(s, s)));
    }

    @Test
    void removingFieldIsBreaking() {
        Schema old = Schema.of(new Field("id", "string"), new Field("age", "int"));
        Schema next = Schema.of(new Field("id", "string"));
        List<Change> changes = SchemaGuard.diff(old, next);
        assertTrue(changes.stream().anyMatch(c -> c.field().equals("age") && c.kind() == ChangeKind.BREAKING));
        assertTrue(SchemaGuard.hasBreaking(changes));
    }

    @Test
    void addingOptionalFieldIsSafe() {
        Schema old = Schema.of(new Field("id", "string"));
        Schema next = Schema.of(new Field("id", "string"), new Field("nickname", "string", false));
        List<Change> changes = SchemaGuard.diff(old, next);
        assertTrue(changes.stream().anyMatch(c -> c.field().equals("nickname") && c.kind() == ChangeKind.SAFE));
        assertFalse(SchemaGuard.hasBreaking(changes));
    }

    @Test
    void addingRequiredFieldIsBreaking() {
        Schema old = Schema.of(new Field("id", "string"));
        Schema next = Schema.of(new Field("id", "string"), new Field("email", "string", true));
        List<Change> changes = SchemaGuard.diff(old, next);
        assertTrue(changes.stream().anyMatch(c -> c.field().equals("email") && c.kind() == ChangeKind.BREAKING));
    }

    @Test
    void typeChangeIsBreaking() {
        Schema old = Schema.of(new Field("age", "int"));
        Schema next = Schema.of(new Field("age", "string"));
        List<Change> changes = SchemaGuard.diff(old, next);
        assertTrue(changes.stream().anyMatch(c -> c.detail().contains("type changed")
                && c.kind() == ChangeKind.BREAKING));
    }

    @Test
    void requiredToOptionalIsSafe() {
        Schema old = Schema.of(new Field("mid", "string", true));
        Schema next = Schema.of(new Field("mid", "string", false));
        List<Change> changes = SchemaGuard.diff(old, next);
        assertEquals(ChangeKind.SAFE, changes.get(0).kind());
        assertFalse(SchemaGuard.hasBreaking(changes));
    }

    @Test
    void optionalToRequiredIsBreaking() {
        Schema old = Schema.of(new Field("mid", "string", false));
        Schema next = Schema.of(new Field("mid", "string", true));
        List<Change> changes = SchemaGuard.diff(old, next);
        assertTrue(changes.stream().anyMatch(c -> c.kind() == ChangeKind.BREAKING));
    }

    @Test
    void breakingSortedFirst() {
        Schema old = Schema.of(new Field("keep", "string", true), new Field("drop", "int"));
        Schema next = Schema.of(new Field("keep", "string", false), new Field("added", "string", true));
        List<Change> changes = SchemaGuard.diff(old, next);
        assertEquals(ChangeKind.BREAKING, changes.get(0).kind());
    }
}
