package models

import "testing"

func TestValidDocumentStatus(t *testing.T) {
	cases := []struct {
		s string
		want bool
	}{
		{DocStatusPending, true},
		{DocStatusIngesting, true},
		{DocStatusReady, true},
		{DocStatusFailed, true},
		{"UNKNOWN", false},
		{"", false},
		{"pending", false}, // case-sensitive
	}
	for _, tc := range cases {
		if got := ValidDocumentStatus(tc.s); got != tc.want {
			t.Errorf("ValidDocumentStatus(%q) = %v, want %v", tc.s, got, tc.want)
		}
	}
}
