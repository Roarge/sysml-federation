// Package assert provides test assertions built on type parameters.
//
// The helpers exist because this repository does not use an empty interface in
// a value position, and the usual assertion libraries take one throughout their
// public API. A generic helper is better than the alternative anyway: the
// compiler rejects a comparison between mismatched types at the call site,
// rather than the test reporting a type mismatch at run time.
//
// Every helper takes a [testing.TB], so the same call works in a test, a
// benchmark and a fuzz target.
package assert

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Equal fails the test unless got and want are equal.
func Equal[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// NotEqual fails the test if got and want are equal.
func NotEqual[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got == want {
		t.Errorf("got %v, want anything else", got)
	}
}

// SliceEqual fails the test unless the two slices have the same elements in the
// same order. It reports the first index that differs rather than dumping both
// slices, which is the part you actually need when one is long.
func SliceEqual[T comparable](t testing.TB, got, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %d elements, want %d\n  got:  %v\n  want: %v", len(got), len(want), got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("element %d: got %v, want %v", i, got[i], want[i])
			return
		}
	}
}

// MapEqual fails the test unless the two maps have the same keys and values.
func MapEqual[K, V comparable](t testing.TB, got, want map[K]V) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %d entries, want %d", len(got), len(want))
		return
	}
	for k, wantV := range want {
		gotV, ok := got[k]
		if !ok {
			t.Errorf("key %v is missing", k)
			return
		}
		if gotV != wantV {
			t.Errorf("key %v: got %v, want %v", k, gotV, wantV)
			return
		}
	}
}

// Len fails the test unless s has exactly n elements. The ~[]E constraint means
// it accepts a named slice type as readily as a plain one.
func Len[S ~[]E, E any](t testing.TB, s S, n int) {
	t.Helper()
	if len(s) != n {
		t.Errorf("got %d elements, want %d: %v", len(s), n, s)
	}
}

// Contains fails the test unless needle is present in haystack. The ~[]E
// constraint accepts a named slice type, matching [Len].
func Contains[S ~[]E, E comparable](t testing.TB, haystack S, needle E) {
	t.Helper()
	for _, v := range haystack {
		if v == needle {
			return
		}
	}
	t.Errorf("%v is not present in %v", needle, haystack)
}

// NoError fails the test if err is non-nil.
func NoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// Error fails the test if err is nil.
func Error(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

// ErrorIs fails the test unless err matches target under [errors.Is], so a
// wrapped sentinel is found at any depth.
func ErrorIs(t testing.TB, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("got error %v, want one matching %v", err, target)
	}
}

// ErrorAs fails the test unless err unwraps to a T, and returns that T so the
// test can go on to assert against its fields. The type parameter is what makes
// this readable: the caller writes the type once, at the call site, instead of
// declaring a variable to pass by address.
// Instantiating T at an interface type makes the assertion vacuous, because
// errors.As matches the first error in the tree against any interface pointer.
// Name a concrete error type.
func ErrorAs[T error](t testing.TB, err error) T {
	t.Helper()
	var target T
	if !errors.As(err, &target) {
		// Fatal rather than Error: the caller is documented to go on and read
		// fields off the result, and a zero value would nil-dereference there.
		t.Fatalf("got error %v, want one of type %T", err, target)
	}
	return target
}

// Must fails the test if err is non-nil, and otherwise returns v. Go cannot
// spread a two-value call beside another argument, so Must takes the value and
// the error separately and the call site reads as two statements,
// `v, err := f()` and then `x := assert.Must(t, v, err)`.
func Must[T any](t testing.TB, v T, err error) T {
	t.Helper()
	if err != nil {
		// Fatalf does not return: it calls runtime.Goexit. There is deliberately
		// no code after it, because none of it could ever run.
		t.Fatalf("unexpected error: %v", err)
	}
	return v
}

// True fails the test unless cond holds, reporting msg.
func True(t testing.TB, cond bool, msg string) {
	t.Helper()
	if !cond {
		t.Errorf("expected %s", msg)
	}
}

// DeepEqual fails the test unless got and want are deeply equal, reporting the
// difference rather than both values in full.
//
// This is the one helper that reaches a value out to a parameter typed as an
// empty interface, inside go-cmp. The type parameter keeps that boundary at the
// edge: a caller still cannot compare two unrelated types by accident.
func DeepEqual[T any](t testing.TB, got, want T, opts ...cmp.Option) {
	t.Helper()
	var (
		diff     string
		failure  string
		panicked bool
	)
	func() {
		// recover returns the empty interface by definition, so it is consumed
		// here rather than stored: the value never reaches a declared type.
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				failure = fmt.Sprint(r)
			}
		}()
		diff = cmp.Diff(want, got, opts...)
	}()
	if panicked {
		t.Errorf("the values could not be compared: %s\n"+
			"go-cmp refuses to read unexported fields. Pass cmp.AllowUnexported or "+
			"cmpopts.EquateErrors, or compare the exported projection instead.", failure)
		return
	}
	if diff != "" {
		t.Errorf("unexpected difference (-want +got):\n%s", diff)
	}
}
