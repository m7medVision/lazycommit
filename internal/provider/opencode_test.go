package provider

import (
	"reflect"
	"testing"
)

func TestOpencodeModelCandidates(t *testing.T) {
	got := opencodeModelCandidates(" opencode/minimax-m2.5-free ", []string{
		"opencode/minimax-m2.5-free",
		" opencode/ling-2.6-flash-free ",
		"",
		"opencode/hy3-preview-free",
	})

	want := []string{
		"opencode/minimax-m2.5-free",
		"opencode/ling-2.6-flash-free",
		"opencode/hy3-preview-free",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
