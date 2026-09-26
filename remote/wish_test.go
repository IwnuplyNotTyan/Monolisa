//go:build ssh

package remote

import (
	"os"
	"testing"
)

func TestEnvDefault(t *testing.T) {
	t.Setenv("MONOLISA_TEST_KEY", "value")
	if got := envDefault("MONOLISA_TEST_KEY", "fallback"); got != "value" {
		t.Errorf("envDefault = %q, want %q", got, "value")
	}

	unsetEnv(t, "MONOLISA_TEST_KEY")
	if got := envDefault("MONOLISA_TEST_KEY", "fallback"); got != "fallback" {
		t.Errorf("envDefault with unset var = %q, want %q", got, "fallback")
	}

	t.Setenv("MONOLISA_TEST_KEY", "")
	if got := envDefault("MONOLISA_TEST_KEY", "fallback"); got != "fallback" {
		t.Errorf("envDefault with empty var = %q, want %q", got, "fallback")
	}
}

func TestServerDefaults(t *testing.T) {
	if _, set := os.LookupEnv("MONOLISA_HOST"); !set && host != "0.0.0.0" {
		t.Errorf("host = %q, want %q", host, "0.0.0.0")
	}
	if _, set := os.LookupEnv("MONOLISA_PORT"); !set && port != "23234" {
		t.Errorf("port = %q, want %q", port, "23234")
	}
}

func TestBannerEmbedded(t *testing.T) {
	if banner == "" {
		t.Error("banner.txt must be embedded into the binary")
	}
}
