package fuzzy

import "testing"

func TestFilter_EmptyQueryReturnsAll(t *testing.T) {
	targets := []string{"build", "test", "clean"}
	got := Filter("", targets)
	if len(got) != len(targets) {
		t.Fatalf("want %d results, got %d", len(targets), len(got))
	}
	for i, m := range got {
		if m.Target != targets[i] {
			t.Errorf("index %d: want %q, got %q", i, targets[i], m.Target)
		}
	}
}

func TestFilter_Matches(t *testing.T) {
	targets := []string{"build", "build-dev", "rebuild", "test", "clean"}
	got := Filter("bld", targets)
	if len(got) == 0 {
		t.Fatal("expected at least one match")
	}
	for _, m := range got {
		if m.Target == "test" || m.Target == "clean" {
			t.Errorf("unexpected match: %q", m.Target)
		}
	}
}

func TestScore_PrefixBeatsInfix(t *testing.T) {
	pre, ok := Score("bui", "build")
	if !ok {
		t.Fatal("prefix match failed")
	}
	inf, ok := Score("bui", "rebuild")
	if !ok {
		t.Fatal("infix match failed")
	}
	if pre.Score >= inf.Score {
		t.Errorf("prefix (%d) should beat infix (%d)", pre.Score, inf.Score)
	}
}

func TestScore_ConsecutiveBeatsScattered(t *testing.T) {
	cons, ok := Score("bui", "build")
	if !ok {
		t.Fatal("consecutive match failed")
	}
	scat, ok := Score("bui", "b_u_i")
	if !ok {
		t.Fatal("scattered match failed")
	}
	if cons.Score >= scat.Score {
		t.Errorf("consecutive (%d) should beat scattered (%d)", cons.Score, scat.Score)
	}
}

func TestScore_CaseInsensitive(t *testing.T) {
	if _, ok := Score("BLD", "build"); !ok {
		t.Error("uppercase query should match lowercase target")
	}
	if _, ok := Score("bld", "BUILD"); !ok {
		t.Error("lowercase query should match uppercase target")
	}
}

func TestScore_NoMatch(t *testing.T) {
	if _, ok := Score("xyz", "build"); ok {
		t.Error("expected no match")
	}
}
