package commands

import (
	"testing"
	"time"
)

func TestBertcheckCooldownFollowsStreamState(t *testing.T) {
	if got := bertcheckCooldown(true); got != 69*time.Second {
		t.Errorf("live cooldown = %s, want 69s", got)
	}
	if got := bertcheckCooldown(false); got != 10*time.Minute {
		t.Errorf("offline cooldown = %s, want 10m", got)
	}
}

func TestPickBertResponseGoldenOnlyWhenAllowed(t *testing.T) {
	available := []string{"beert", "goldenzazabert"}

	for i := 0; i < 200000; i++ {
		if got := pickBertResponse("someone", available, "sir_lysergium", false); got == "goldenzazabert" {
			t.Fatal("goldenzazabert was pulled while offline")
		}
	}
}

func TestPickBertResponseSkipsGoldenzazabertRecord(t *testing.T) {
	if got := pickBertResponse("someone", []string{"goldenzazabert"}, "sir_lysergium", false); got != "" {
		t.Errorf("response = %q, want no response when only goldenzazabert is available", got)
	}
}
