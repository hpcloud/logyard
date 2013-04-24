package drain

import (
	"testing"
)

func TestEqual(t *testing.T) {
	m1 := map[string]string{"foo": "bar", "a": "b"}
	m2 := map[string]string{"foo": "bar", "a": "b"}
	if len(MapDiff(m1, m2)) != 0 {
		t.Fatal("not equal")
	}
}

func TestAdded(t *testing.T) {
	m1 := map[string]string{"a": "b", "foo": "bar"}
	m2 := map[string]string{"a": "b", "cake": "x", "foo": "bar"}
	c := MapDiff(m1, m2)
	if len(c) != 1 {
		t.Fatalf("not a single change: %+v", c)
	}
	if c[0].Deleted {
		t.Fatal("a deletion")
	}
	if c[0].Key != "cake" {
		t.Fatal("invalid key")
	}
	if c[0].NewValue != "x" {
		t.Fatal("invalid value")
	}
}

func TestChanged(t *testing.T) {
	m1 := map[string]string{"a": "b", "cake": "x", "foo": "bar"}
	m2 := map[string]string{"a": "b", "cake": "y", "foo": "bar"}
	c := MapDiff(m1, m2)
	if len(c) != 1 {
		t.Fatalf("not a single change: %+v", c)
	}
	if c[0].Deleted {
		t.Fatal("a deletion")
	}
	if c[0].Key != "cake" {
		t.Fatal("invalid key")
	}
	if c[0].OldValue != "x" || c[0].NewValue != "y" {
		t.Fatal("invalid value")
	}
}

func TestDeleted(t *testing.T) {
	m1 := map[string]string{"a": "b", "cake": "x", "foo": "bar"}
	m2 := map[string]string{"a": "b", "cake": "x"}
	c := MapDiff(m1, m2)
	if len(c) != 1 {
		t.Fatalf("not a single change: %+v", c)
	}
	if !c[0].Deleted {
		t.Fatal("not a deletion")
	}
	if c[0].Key != "foo" {
		t.Fatal("invalid key")
	}
}
