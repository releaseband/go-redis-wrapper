package internal

import (
	"errors"
	"testing"
)

func TestCase(t *testing.T, name string) func(fail, success func(t *testing.T)) {
	return func(fail, success func(t *testing.T)) {
		t.Run(name, func(t *testing.T) {
			t.Run("fail", fail)
			t.Run("success", success)
		})
	}
}

func ErrorChecker(t *testing.T) func(expErr, gotErr error) {
	return func(expErr, gotErr error) {
		if gotErr == nil && expErr == nil {
			return
		}

		if gotErr == nil && expErr != nil {
			t.Fatal("the error received must not be nil")
		}

		if gotErr != nil && expErr == nil {
			t.Fatalf("gotErr := '%s'; the error received must be nil", errors.Unwrap(gotErr))
		}

		if !errors.Is(gotErr, expErr) {
			t.Fatalf("expErr := '%s' ; gotErr := '%s': gotErr should must be equal expErr",
				expErr.Error(), errors.Unwrap(gotErr))
		}
	}
}
