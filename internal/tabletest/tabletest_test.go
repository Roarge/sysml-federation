package tabletest_test

import (
	"strings"
	"testing"

	"github.com/Roarge/sysml-federation/internal/tabletest"
)

func TestRunExercisesEveryCase(t *testing.T) {
	seen := map[string]bool{}

	tabletest.Run(t, []tabletest.Case[string, int]{
		{Name: "empty", In: "", Want: 0},
		{Name: "one word", In: "part", Want: 1},
		{Name: "three words", In: "part requirement stage", Want: 3},
	}, func(t *testing.T, in string) int {
		seen[in] = true
		return len(strings.Fields(in))
	})

	if len(seen) != 3 {
		t.Fatalf("ran %d cases, want 3", len(seen))
	}
}

func TestRunErrExercisesEveryCase(t *testing.T) {
	count := 0

	tabletest.RunErr(t, []tabletest.ErrCase[int, string]{
		{Name: "positive", In: 2, Want: "ok"},
		{Name: "negative", In: -1, WantErr: true},
	}, func(t *testing.T, in int) (string, error) {
		count++
		if in < 0 {
			return "", errNegative
		}
		return "ok", nil
	})

	if count != 2 {
		t.Fatalf("ran %d cases, want 2", count)
	}
}
