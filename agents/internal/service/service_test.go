package service

import "testing"

func TestShouldSucceed_IsDeterministic(t *testing.T) {
	a := shouldSucceed("scan-1", "https://example.com", "prompt_injection")
	b := shouldSucceed("scan-1", "https://example.com", "prompt_injection")
	if a != b {
		t.Fatalf("non-deterministic result: %v vs %v", a, b)
	}
}

func TestBuildPrompt_HasModePrefix(t *testing.T) {
	got := buildPrompt("jailbreak", "red_team")
	if got == "" || got[0] != '[' {
		t.Fatalf("expected mode prefix, got %q", got)
	}
}
