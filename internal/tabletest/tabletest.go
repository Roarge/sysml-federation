// Package tabletest runs table-driven tests without the boilerplate that
// usually surrounds them.
//
// The table-driven test is the dominant shape in Go, and every one of them
// repeats the same loop, the same t.Run, and the same comparison. Type
// parameters let that be written once: the input and output types are stated at
// the call site, so a case whose Want does not match what the function returns
// is a compile error rather than a run-time surprise.
package tabletest

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// Case is one row of a table: an input, the expected output, and a name that
// becomes the subtest name.
type Case[In, Out any] struct {
	Name string
	In   In
	Want Out
	// Opts are passed to the comparison, for types needing an exported-field
	// or floating-point tolerance option.
	Opts []cmp.Option
}

// Run executes fn against every case as a named subtest and compares the result
// with Want.
func Run[In, Out any](t *testing.T, cases []Case[In, Out], fn func(*testing.T, In) Out) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()
			assert.DeepEqual(t, fn(t, tc.In), tc.Want, tc.Opts...)
		})
	}
}

// ErrCase is a row for a function returning a value and an error.
type ErrCase[In, Out any] struct {
	Name string
	In   In
	Want Out
	// WantErr states that the call must fail. When set, Want is not compared.
	WantErr bool
	// ErrIs, when set, must match the returned error under errors.Is.
	ErrIs error
	Opts  []cmp.Option
}

// RunErr executes fn against every case, asserting the error expectation first
// and the value only when no error was expected.
func RunErr[In, Out any](
	t *testing.T,
	cases []ErrCase[In, Out],
	fn func(*testing.T, In) (Out, error),
) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()
			got, err := fn(t, tc.In)
			if tc.WantErr || tc.ErrIs != nil {
				assert.Error(t, err)
				if tc.ErrIs != nil {
					assert.ErrorIs(t, err, tc.ErrIs)
				}
				return
			}
			if err != nil {
				// Reporting the comparison as well would add a second failure
				// about a zero value, which says nothing about the real fault.
				t.Errorf("unexpected error: %v", err)
				return
			}
			assert.DeepEqual(t, got, tc.Want, tc.Opts...)
		})
	}
}
