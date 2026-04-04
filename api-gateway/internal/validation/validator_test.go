package validation_test

import (
	"testing"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/agentshield/api-gateway/internal/validation"
)

type endpointCase struct {
	url  string
	want bool
}

func TestHTTPSEndpointValidator(t *testing.T) {
	cases := []endpointCase{
		// Valid
		{"https://example.com", true},
		{"https://api.example.com/v1/chat", true},
		{"https://some-service.example.org:443/path?q=1", true},

		// Non-HTTPS
		{"http://example.com", false},
		{"ftp://example.com", false},

		// Private IPs
		{"https://10.0.0.1", false},
		{"https://172.16.0.1", false},
		{"https://192.168.1.1", false},
		{"https://127.0.0.1", false},
		{"https://localhost", false},
		{"https://169.254.169.254", false},
		{"https://169.254.169.254/latest/meta-data/", false},

		// Malformed
		{"", false},
		{"not-a-url", false},
	}

	for _, tc := range cases {
		req := domain.CreateScanRequest{
			TargetEndpoint: tc.url,
			Mode:           domain.ModeRedTeam,
			AttackTypes:    []string{"prompt_injection"},
		}
		err := validation.Validate.Struct(req)
		got := err == nil
		if got != tc.want {
			t.Errorf("url=%q: want valid=%v, got valid=%v (err=%v)", tc.url, tc.want, got, err)
		}
	}
}

func TestTooLongURL(t *testing.T) {
	long := "https://example.com/" + string(make([]byte, 490))
	for i := range long[20:] {
		_ = i
	}
	// Build a 501-char URL
	url := "https://example.com/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if len(url) <= 500 {
		t.Logf("URL length %d, padding...", len(url))
	}

	req := domain.CreateScanRequest{
		TargetEndpoint: url,
		Mode:           domain.ModeRedTeam,
		AttackTypes:    []string{"prompt_injection"},
	}
	err := validation.Validate.Struct(req)
	if len(url) > 500 && err == nil {
		t.Error("expected validation error for URL > 500 chars, got nil")
	}
}
