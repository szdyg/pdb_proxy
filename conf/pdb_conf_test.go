package conf

import (
	"testing"
	"time"
)

func TestGetDurationEnvWithDefault(t *testing.T) {
	const key = "PDB_PROXY_TEST_CACHE_TTL"
	const fallback = 3 * time.Hour

	t.Run("uses default when unset", func(t *testing.T) {
		t.Setenv(key, "")
		if got := GetDurationEnvWithDefault(key, fallback); got != fallback {
			t.Fatalf("got %s, want %s", got, fallback)
		}
	})

	t.Run("parses duration", func(t *testing.T) {
		t.Setenv(key, "30m")
		if got := GetDurationEnvWithDefault(key, fallback); got != 30*time.Minute {
			t.Fatalf("got %s, want 30m", got)
		}
	})

	t.Run("zero means forever", func(t *testing.T) {
		t.Setenv(key, "0")
		if got := GetDurationEnvWithDefault(key, fallback); got != 0 {
			t.Fatalf("got %s, want 0", got)
		}
	})

	for _, value := range []string{"invalid", "-1h"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(key, value)
			if got := GetDurationEnvWithDefault(key, fallback); got != fallback {
				t.Fatalf("got %s, want %s", got, fallback)
			}
		})
	}
}
