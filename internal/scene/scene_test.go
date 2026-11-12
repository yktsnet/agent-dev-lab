package scene

import (
	"reflect"
	"strings"
	"testing"
)

func TestGenerateIsDeterministic(t *testing.T) {
	a := Generate(42, 50, Overlap{Collide: 2, Prefix: 2})
	b := Generate(42, 50, Overlap{Collide: 2, Prefix: 2})
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed produced different scenes")
	}
}

func TestOverlapCounts(t *testing.T) {
	for _, ov := range []Overlap{{}, {Collide: 3}, {Prefix: 3}, {Collide: 2, Prefix: 4}} {
		sc := Generate(7, 60, ov)
		seen := map[string]int{}
		for _, s := range sc {
			seen[s.ID]++
		}
		dup := 0
		for _, n := range seen {
			dup += n - 1
		}
		if dup != ov.Collide {
			t.Errorf("%+v: duplicated IDs = %d, want %d", ov, dup, ov.Collide)
		}
		traps := 0
		for _, s := range sc {
			for _, other := range sc {
				if other.ID != s.ID && strings.HasPrefix(other.ID, s.ID) && strings.Contains(s.Expect[1].SQL, "LIKE") {
					traps++
				}
			}
		}
		if traps != ov.Prefix {
			t.Errorf("%+v: prefix traps = %d, want %d", ov, traps, ov.Prefix)
		}
	}
}
