package model

import (
	"testing"
)

func TestStatusString(t *testing.T) {
	tests := []struct{
		stat Status
		want string
	}{
		{Unknown, ""},
		{Unmarked, ""},
		{Pending, "!"},
		{Cleared, "*"},
	}
	for _, test := range tests {
		if got, want := test.stat.String(), test.want; got != want {
			t.Errorf("%v.String()=%q want %q", test.stat, got, want)
		}
	}
}