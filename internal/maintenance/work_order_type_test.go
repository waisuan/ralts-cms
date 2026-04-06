package maintenance

import (
	"strings"
	"testing"
)

func TestStandardWorkOrderTypes_OrderAndUniqueness(t *testing.T) {
	t.Parallel()
	if len(StandardWorkOrderTypes) == 0 {
		t.Fatal("StandardWorkOrderTypes must be non-empty")
	}
	seen := make(map[string]struct{}, len(StandardWorkOrderTypes))
	for _, s := range StandardWorkOrderTypes {
		if _, dup := seen[s]; dup {
			t.Fatalf("duplicate standard type %q", s)
		}
		seen[s] = struct{}{}
	}
}

func TestIsStandardWorkOrderType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    string
		want bool
	}{
		{"Preventive", true},
		{"Corrective", true},
		{"Emergency", true},
		{"Inspection", true},
		{"", false},
		{"Other", false},
		{"something custom", false},
		{"preventive", false},
	}
	for _, tc := range cases {
		if got := IsStandardWorkOrderType(tc.s); got != tc.want {
			t.Errorf("IsStandardWorkOrderType(%q) = %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestCountByWorkOrderTypeSQL_IncludesEveryStandardType(t *testing.T) {
	t.Parallel()
	q := CountByWorkOrderTypeSQL()
	for _, typ := range StandardWorkOrderTypes {
		if !strings.Contains(q, typ) {
			t.Errorf("query missing standard type %q", typ)
		}
	}
	if !strings.Contains(q, `"serialNumber" = $1`) {
		t.Error("query missing machine filter placeholder")
	}
}

func TestCountByWorkOrderTypeSQL_StandardTypeCaseOrder(t *testing.T) {
	t.Parallel()
	q := CountByWorkOrderTypeSQL()
	pos := 0
	for _, typ := range StandardWorkOrderTypes {
		frag := `COUNT(CASE WHEN "workOrderType" = '` + typ + `' THEN 1 END)`
		idx := strings.Index(q[pos:], frag)
		if idx < 0 {
			t.Fatalf("query missing ordered fragment for %q after offset %d", typ, pos)
		}
		pos += idx + len(frag)
	}
}
