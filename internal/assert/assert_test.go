package assert_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// recorder stands in for *testing.T so the helpers can be tested for the
// failing case as well as the passing one. Without it a test could only ever
// assert that a passing assertion passes, which is the half that never breaks.
//
// The embedded TB is the real one, so a helper that reaches a method this type
// does not override delegates rather than dereferencing nil.
type recorder struct {
	testing.TB
	failed bool
	msgs   []string
}

func (r *recorder) Helper() {}

// The variadic empty interface is imposed by testing.TB, which this type has
// to satisfy. It is the signature of the standard library, not a choice here.
func (r *recorder) Errorf(format string, args ...any) { //nointerface:allow
	r.failed = true
	r.msgs = append(r.msgs, fmt.Sprintf(format, args...))
}
func (r *recorder) Fatalf(format string, args ...any) { //nointerface:allow
	r.Errorf(format, args...)
}

func TestEqualPasses(t *testing.T) {
	r := &recorder{TB: t}
	assert.Equal(r, 42, 42)
	assert.Equal(r, "sysml", "sysml")
	if r.failed {
		t.Fatalf("Equal reported a failure for equal values: %v", r.msgs)
	}
}

func TestEqualFails(t *testing.T) {
	r := &recorder{TB: t}
	assert.Equal(r, 41, 42)
	if !r.failed {
		t.Fatal("Equal accepted 41 == 42")
	}
}

func TestNotEqual(t *testing.T) {
	r := &recorder{TB: t}
	assert.NotEqual(r, 1, 2)
	if r.failed {
		t.Fatalf("NotEqual reported a failure for differing values: %v", r.msgs)
	}
	r2 := &recorder{TB: t}
	assert.NotEqual(r2, 3, 3)
	if !r2.failed {
		t.Fatal("NotEqual accepted 3 != 3")
	}
}

func TestSliceEqual(t *testing.T) {
	r := &recorder{TB: t}
	assert.SliceEqual(r, []int{1, 2, 3}, []int{1, 2, 3})
	if r.failed {
		t.Fatalf("SliceEqual reported a failure for identical slices: %v", r.msgs)
	}

	for name, tc := range map[string]struct{ got, want []int }{
		"differing length":  {[]int{1, 2}, []int{1, 2, 3}},
		"differing element": {[]int{1, 9, 3}, []int{1, 2, 3}},
	} {
		t.Run(name, func(t *testing.T) {
			r := &recorder{TB: t}
			assert.SliceEqual(r, tc.got, tc.want)
			if !r.failed {
				t.Fatalf("SliceEqual accepted %v == %v", tc.got, tc.want)
			}
		})
	}
}

func TestMapEqual(t *testing.T) {
	r := &recorder{TB: t}
	assert.MapEqual(r, map[string]int{"a": 1, "b": 2}, map[string]int{"b": 2, "a": 1})
	if r.failed {
		t.Fatalf("MapEqual reported a failure for identical maps: %v", r.msgs)
	}

	for name, tc := range map[string]struct{ got, want map[string]int }{
		"differing size":  {map[string]int{"a": 1}, map[string]int{"a": 1, "b": 2}},
		"missing key":     {map[string]int{"a": 1, "c": 2}, map[string]int{"a": 1, "b": 2}},
		"differing value": {map[string]int{"a": 1, "b": 9}, map[string]int{"a": 1, "b": 2}},
	} {
		t.Run(name, func(t *testing.T) {
			r := &recorder{TB: t}
			assert.MapEqual(r, tc.got, tc.want)
			if !r.failed {
				t.Fatalf("MapEqual accepted %v == %v", tc.got, tc.want)
			}
		})
	}
}

func TestLen(t *testing.T) {
	r := &recorder{TB: t}
	assert.Len(r, []string{"a", "b"}, 2)
	if r.failed {
		t.Fatalf("Len reported a failure for a correct length: %v", r.msgs)
	}
	r2 := &recorder{TB: t}
	assert.Len(r2, []string{"a"}, 2)
	if !r2.failed {
		t.Fatal("Len accepted a length of 1 as 2")
	}
}

func TestContains(t *testing.T) {
	r := &recorder{TB: t}
	assert.Contains(r, []string{"part", "requirement"}, "requirement")
	if r.failed {
		t.Fatalf("Contains missed a present element: %v", r.msgs)
	}
	r2 := &recorder{TB: t}
	assert.Contains(r2, []string{"part"}, "requirement")
	if !r2.failed {
		t.Fatal("Contains found an absent element")
	}
}

var errSentinel = errors.New("sentinel")

func TestNoError(t *testing.T) {
	r := &recorder{TB: t}
	assert.NoError(r, nil)
	if r.failed {
		t.Fatalf("NoError reported a failure for nil: %v", r.msgs)
	}
	r2 := &recorder{TB: t}
	assert.NoError(r2, errSentinel)
	if !r2.failed {
		t.Fatal("NoError accepted a non-nil error")
	}
}

func TestError(t *testing.T) {
	r := &recorder{TB: t}
	assert.Error(r, errSentinel)
	if r.failed {
		t.Fatalf("Error reported a failure for a non-nil error: %v", r.msgs)
	}
	r2 := &recorder{TB: t}
	assert.Error(r2, nil)
	if !r2.failed {
		t.Fatal("Error accepted nil as a failure")
	}
}

func TestErrorIs(t *testing.T) {
	wrapped := fmt.Errorf("reading model: %w", errSentinel)
	r := &recorder{TB: t}
	assert.ErrorIs(r, wrapped, errSentinel)
	if r.failed {
		t.Fatalf("ErrorIs missed a wrapped sentinel: %v", r.msgs)
	}
	r2 := &recorder{TB: t}
	assert.ErrorIs(r2, errors.New("unrelated"), errSentinel)
	if !r2.failed {
		t.Fatal("ErrorIs matched an unrelated error")
	}
}

type notFoundError struct{ id string }

func (e *notFoundError) Error() string { return "not found: " + e.id }

func TestErrorAs(t *testing.T) {
	wrapped := fmt.Errorf("resolving: %w", &notFoundError{id: "R-1"})
	r := &recorder{TB: t}
	got := assert.ErrorAs[*notFoundError](r, wrapped)
	if r.failed {
		t.Fatalf("ErrorAs missed a wrapped typed error: %v", r.msgs)
	}
	if got == nil || got.id != "R-1" {
		t.Fatalf("ErrorAs returned %#v, want the *notFoundError carrying R-1", got)
	}

	// An error of another type. The real *testing.T would stop the test here,
	// but the recorder's Fatalf only records, so the zero value comes back.
	r2 := &recorder{TB: t}
	other := assert.ErrorAs[*notFoundError](r2, errSentinel)
	if !r2.failed {
		t.Fatal("ErrorAs matched an error of another type")
	}
	if other != nil {
		t.Fatalf("ErrorAs returned %#v after failing, want the zero value", other)
	}
}

func TestMust(t *testing.T) {
	r := &recorder{TB: t}
	got := assert.Must(r, 7, nil)
	if r.failed || got != 7 {
		t.Fatalf("Must(7, nil) = %v, failed=%v", got, r.failed)
	}
	r2 := &recorder{TB: t}
	assert.Must(r2, 0, errSentinel)
	if !r2.failed {
		t.Fatal("Must accepted a non-nil error")
	}
}

func TestTrue(t *testing.T) {
	r := &recorder{TB: t}
	assert.True(r, true, "the model has a root")
	if r.failed {
		t.Fatalf("True reported a failure for a condition that holds: %v", r.msgs)
	}
	r2 := &recorder{TB: t}
	assert.True(r2, false, "the model has a root")
	if !r2.failed {
		t.Fatal("True accepted a condition that does not hold")
	}
	if len(r2.msgs) != 1 || r2.msgs[0] != "expected the model has a root" {
		t.Fatalf("True reported %v, want the message it was given", r2.msgs)
	}
}

type stage struct {
	Name     string
	Capacity int
}

func TestDeepEqual(t *testing.T) {
	r := &recorder{TB: t}
	assert.DeepEqual(r, stage{"parse", 100}, stage{"parse", 100})
	if r.failed {
		t.Fatalf("DeepEqual reported a failure for identical structs: %v", r.msgs)
	}
	r2 := &recorder{TB: t}
	assert.DeepEqual(r2, stage{"parse", 100}, stage{"parse", 90})
	if !r2.failed {
		t.Fatal("DeepEqual accepted differing structs")
	}
}

// --- regressions ----------------------------------------------------------

type withUnexported struct{ id string }

func TestDeepEqualDoesNotPanicOnUnexportedFields(t *testing.T) {
	// go-cmp panics rather than returning a difference when it meets an
	// unexported field. A test helper must report that, not take the process
	// down with it.
	r := &recorder{TB: t}
	assert.DeepEqual(r, withUnexported{"a"}, withUnexported{"b"})
	if !r.failed {
		t.Fatal("DeepEqual accepted values it could not compare")
	}
}

func TestDeepEqualDoesNotPanicOnErrors(t *testing.T) {
	r := &recorder{TB: t}
	assert.DeepEqual(r, errors.New("one"), errors.New("two"))
	if !r.failed {
		t.Fatal("DeepEqual accepted two errors it could not compare")
	}
}

func TestContainsAcceptsANamedSliceType(t *testing.T) {
	type names []string
	r := &recorder{TB: t}
	assert.Contains(r, names{"part", "requirement"}, "requirement")
	if r.failed {
		t.Fatalf("Contains rejected a named slice type: %v", r.msgs)
	}
}
