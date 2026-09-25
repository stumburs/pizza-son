package services

import (
	"path/filepath"
	"pizza-son/internal/store"
	"testing"
)

func bertTestData() map[string]*ChannelData {
	return map[string]*ChannelData{
		"sir_lysergium": {
			Berts: []string{"bertnard", "beert"},
			UserStats: map[string]UserStats{
				"alice": {
					TotalActivations: 10,
					BertRecords: map[string]BertRecord{
						"bertnard": {Count: 8},
						"gonebert": {Count: 2},
					},
				},
				"bob": {
					TotalActivations: 5,
					BertRecords: map[string]BertRecord{
						"beert": {Count: 5},
					},
				},
			},
		},
		"123456789012345678": {
			Berts: []string{"<a:bert:1>"},
			UserStats: map[string]UserStats{
				"alice": {
					TotalActivations: 7,
					BertRecords: map[string]BertRecord{
						"<a:bert:1>": {Count: 7},
					},
				},
			},
		},
	}
}

func TestGetUserStatsTotalsMatchRecordedActivations(t *testing.T) {
	s := &BertService{Data: bertTestData()}

	stats := s.GetUserStats("sir_lysergium", "alice")
	if stats.TotalBertchecks != 10 {
		t.Errorf("TotalBertchecks = %d, want 10 (every recorded activation)", stats.TotalBertchecks)
	}
	if stats.ChannelTotalBertchecks != 15 {
		t.Errorf("ChannelTotalBertchecks = %d, want 15 (every recorded activation)", stats.ChannelTotalBertchecks)
	}
	if stats.TotalBerts != 2 {
		t.Errorf("TotalBerts = %d, want 2", stats.TotalBerts)
	}
	if stats.BertsCollectedOutOfAll != 1 {
		t.Errorf("BertsCollectedOutOfAll = %d, want 1 (gonebert is no longer active)", stats.BertsCollectedOutOfAll)
	}
	if stats.MostCommonBert != "bertnard" || stats.MostCommonCount != 8 {
		t.Errorf("most common = %s (%dx), want bertnard (8x)", stats.MostCommonBert, stats.MostCommonCount)
	}

	unknown := s.GetUserStats("sir_lysergium", "carol")
	if unknown.TotalBertchecks != 0 {
		t.Errorf("unknown user TotalBertchecks = %d, want 0", unknown.TotalBertchecks)
	}
	if unknown.ChannelTotalBertchecks != 15 || unknown.TotalBerts != 2 {
		t.Errorf("unknown user channel stats = %d/%d, want 15/2", unknown.ChannelTotalBertchecks, unknown.TotalBerts)
	}
}

func TestGlobalTotalExcludesDiscordChannels(t *testing.T) {
	s := &BertService{Data: bertTestData()}

	if got := s.GlobalTotalLocked(); got != 15 {
		t.Errorf("GlobalTotalLocked() = %d, want 15 (Discord channels excluded)", got)
	}
}

func TestRegisterActivationReturnsSiteGlobalTotal(t *testing.T) {
	st := store.New(filepath.Join(t.TempDir(), "stats.json"), &map[string]*ChannelData{})
	if err := st.EnsureDir(); err != nil {
		t.Fatal(err)
	}
	s := &BertService{store: st, Data: *st.Data()}

	// Discord activations are not part of the global total the site reports
	if got := s.RegisterActivation("123456789012345678", "alice", "<a:bert:1>", false, false, false); got != 0 {
		t.Errorf("discord activation returned %d, want 0", got)
	}
	if got := s.RegisterActivation("sir_lysergium", "alice", "bertnard", false, false, false); got != 1 {
		t.Errorf("first activation returned %d, want 1", got)
	}
	if got := s.RegisterActivation("sir_lysergium", "bob", "beert", false, false, false); got != 2 {
		t.Errorf("second activation returned %d, want 2", got)
	}
	if got := s.GlobalTotalLocked(); got != 2 {
		t.Errorf("GlobalTotalLocked() = %d, want 2", got)
	}
}
