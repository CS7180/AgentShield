package validation

import (
	"net"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
	_ = Validate.RegisterValidation("https_endpoint", validateHTTPSEndpoint)
}

// validateHTTPSEndpoint rejects:
//   - Non-HTTPS schemes
//   - Private/reserved IPs (RFC1918, loopback, link-local, 169.254.169.254)
//   - URLs longer than 500 characters
func validateHTTPSEndpoint(fl validator.FieldLevel) bool {
	raw := fl.Field().String()
	if len(raw) > 500 {
		return false
	}

	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}

	if u.Scheme != "https" {
		return false
	}

	hostname := u.Hostname()
	if isPrivateHost(hostname) {
		return false
	}

	return true
}

var privateRanges = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"127.0.0.0/8",
	"::1/128",
	"169.254.0.0/16",
	"fc00::/7",
}

func isPrivateHost(hostname string) bool {
	// Block AWS metadata IP directly
	if hostname == "169.254.169.254" {
		return true
	}

	// Block localhost
	if strings.EqualFold(hostname, "localhost") {
		return true
	}

	ip := net.ParseIP(hostname)
	if ip == nil {
		// It's a hostname, not an IP — allow (DNS resolution happens later)
		return false
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
