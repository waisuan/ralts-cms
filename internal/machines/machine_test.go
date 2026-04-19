package machines

import (
	"testing"
	"time"
)

func TestSetTimestamps(t *testing.T) {
	t.Parallel()
	t.Run("new machine", func(t *testing.T) {
		t.Parallel()
		machine := &Machine{SerialNumber: "TEST123"}
		machine.SetTimestamps()
		if machine.CreatedAt.IsZero() || machine.UpdatedAt.IsZero() {
			t.Fatal("expected non-zero CreatedAt and UpdatedAt")
		}
		if !machine.CreatedAt.Equal(machine.UpdatedAt) {
			t.Fatalf("CreatedAt and UpdatedAt should match on first set: %v vs %v", machine.CreatedAt, machine.UpdatedAt)
		}
	})
	t.Run("existing machine", func(t *testing.T) {
		t.Parallel()
		originalCreatedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		machine := &Machine{
			SerialNumber: "TEST123",
			CreatedAt:    originalCreatedAt,
			UpdatedAt:    originalCreatedAt,
		}
		machine.SetTimestamps()
		if !machine.CreatedAt.Equal(originalCreatedAt) {
			t.Fatalf("CreatedAt changed: %v", machine.CreatedAt)
		}
		if !machine.UpdatedAt.After(originalCreatedAt) {
			t.Fatalf("UpdatedAt should advance: %v", machine.UpdatedAt)
		}
	})
}

func TestIsPpmDateUnset(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		tm   time.Time
		want bool
	}{
		{"zero time", time.Time{}, true},
		{"year 1 Jan 1 UTC", time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{"year 1 Dec 31 UTC", time.Date(1, 12, 31, 0, 0, 0, 0, time.UTC), true},
		{"year 2", time.Date(2, 1, 1, 0, 0, 0, 0, time.UTC), false},
		{"normal", time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsPpmDateUnset(tc.tm); got != tc.want {
				t.Fatalf("IsPpmDateUnset(%v) = %v, want %v", tc.tm, got, tc.want)
			}
		})
	}
}
