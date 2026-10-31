package services

import (
	"path/filepath"
	"pizza-son/internal/store"
	"testing"
	"time"
)

func purgeTestChannel() *ChannelData {
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)
	return &ChannelData{
		Berts: []string{"bertnard", "beert", "goldenzazabert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 3,
				TotalZazas:       2,
				TotalZazaLs:      2,
				TotalDoubleZazas: 1,
				BertRecords: map[string]BertRecord{
					"bertnard": {
						Count:      2,
						ZazaCount:  1,
						ZazaLCount: 1,
						FirstSeen:  base,
						LastSeen:   base.Add(90 * time.Minute),
					},
					"beert": {
						Count:           1,
						ZazaCount:       1,
						ZazaLCount:      1,
						DoubleZazaCount: 1,
						FirstSeen:       base.Add(30 * time.Minute),
						LastSeen:        base.Add(30 * time.Minute),
					},
				},
				DailyActivations: map[string]int{bertDayKey(base): 3},
				HourlyActivations: map[string]int{
					bertHourKey(base):                2,
					bertHourKey(base.Add(time.Hour)): 1,
				},
			},
			"bob": {
				TotalActivations: 1,
				BertRecords: map[string]BertRecord{
					"beert": {
						Count:     1,
						FirstSeen: base.Add(15 * time.Minute),
						LastSeen:  base.Add(15 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 1},
				HourlyActivations: map[string]int{bertHourKey(base): 1},
			},
		},
		DailyActivations: map[string]int{bertDayKey(base): 4},
		HourlyActivations: map[string]int{
			bertHourKey(base):                3,
			bertHourKey(base.Add(time.Hour)): 1,
		},
		GoldenzazabertEvents: []GoldenzazabertEvent{
			{Username: "alice", Timestamp: base.Add(30 * time.Minute)},
			{Username: "bob", Timestamp: base.Add(15 * time.Minute)},
		},
		ActivationLog: []BertActivation{
			{Username: "alice", Bert: "bertnard", Timestamp: base, IsZaza: true},
			{Username: "bob", Bert: "beert", Timestamp: base.Add(15 * time.Minute)},
			{Username: "alice", Bert: "beert", Timestamp: base.Add(30 * time.Minute), IsZaza: true, IsZazaL: true, DoubleZaza: true},
			{Username: "alice", Bert: "bertnard", Timestamp: base.Add(90 * time.Minute), IsZazaL: true},
		},
	}
}

func TestPurgeUserActivationsAdjustsEveryAggregate(t *testing.T) {
	ch := purgeTestChannel()
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)

	result, err := PurgeUserActivations(ch, PurgeOptions{
		User: "alice",
		From: base.Add(20 * time.Minute),
		To:   base.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Removed != 2 {
		t.Fatalf("Removed = %d, want 2", result.Removed)
	}
	if result.RemovedByBert["bertnard"] != 1 || result.RemovedByBert["beert"] != 1 {
		t.Errorf("RemovedByBert = %v, want one of each", result.RemovedByBert)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", result.Warnings)
	}

	alice := ch.UserStats["alice"]
	if alice.TotalActivations != 1 || alice.TotalZazas != 1 || alice.TotalZazaLs != 0 || alice.TotalDoubleZazas != 0 {
		t.Errorf("alice totals = %d/%d/%d/%d, want 1/1/0/0",
			alice.TotalActivations, alice.TotalZazas, alice.TotalZazaLs, alice.TotalDoubleZazas)
	}

	bertnard := alice.BertRecords["bertnard"]
	if bertnard.Count != 1 || bertnard.ZazaCount != 1 || bertnard.ZazaLCount != 0 || bertnard.DoubleZazaCount != 0 {
		t.Errorf("bertnard record = %+v, want one plain-ish activation left", bertnard)
	}
	if !bertnard.LastSeen.Equal(base) {
		t.Errorf("bertnard.LastSeen = %s, want %s (latest remaining activation)", bertnard.LastSeen, base)
	}
	if _, exists := alice.BertRecords["beert"]; exists {
		t.Error("beert record should be removed when its only activation is purged")
	}

	if got := ch.DailyActivations[bertDayKey(base)]; got != 2 {
		t.Errorf("channel daily = %d, want 2", got)
	}
	if got := ch.HourlyActivations[bertHourKey(base)]; got != 2 {
		t.Errorf("channel 12:00 hour = %d, want 2", got)
	}
	if _, exists := ch.HourlyActivations[bertHourKey(base.Add(time.Hour))]; exists {
		t.Error("empty channel 13:00 hour should be deleted")
	}
	if got := alice.DailyActivations[bertDayKey(base)]; got != 1 {
		t.Errorf("alice daily = %d, want 1", got)
	}

	if len(ch.ActivationLog) != 2 {
		t.Fatalf("activation log has %d entries, want 2", len(ch.ActivationLog))
	}
	for _, activation := range ch.ActivationLog {
		if activation.Username == "alice" && activation.Timestamp.After(base) {
			t.Errorf("alice activation at %s was not purged", activation.Timestamp)
		}
	}

	if len(ch.GoldenzazabertEvents) != 1 || ch.GoldenzazabertEvents[0].Username != "bob" {
		t.Errorf("goldenzazabert events = %+v, want only bob's", ch.GoldenzazabertEvents)
	}
}

func TestPurgeUserActivationsKeepsLaterActivations(t *testing.T) {
	ch := purgeTestChannel()
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)

	// Remove only alice's middle activation; her 13:00 activation must still
	// be reflected in the record.
	_, err := PurgeUserActivations(ch, PurgeOptions{
		User: "alice",
		From: base.Add(20 * time.Minute),
		To:   base.Add(40 * time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}

	alice := ch.UserStats["alice"]
	if alice.TotalActivations != 2 {
		t.Errorf("alice total = %d, want 2", alice.TotalActivations)
	}
	bertnard := alice.BertRecords["bertnard"]
	if bertnard.Count != 2 {
		t.Errorf("bertnard count = %d, want 2", bertnard.Count)
	}
	if !bertnard.LastSeen.Equal(base.Add(90 * time.Minute)) {
		t.Errorf("bertnard.LastSeen = %s, want %s", bertnard.LastSeen, base.Add(90*time.Minute))
	}
	if alice.TotalZazas != 1 || alice.TotalZazaLs != 1 || alice.TotalDoubleZazas != 0 {
		t.Errorf("alice zaza totals = %d/%d/%d, want 1/1/0",
			alice.TotalZazas, alice.TotalZazaLs, alice.TotalDoubleZazas)
	}
}

func TestPurgeUserActivationsRefusesUnloggedRange(t *testing.T) {
	ch := purgeTestChannel()
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)
	before := ch.UserStats["alice"].TotalActivations

	_, err := PurgeUserActivations(ch, PurgeOptions{
		User: "alice",
		From: base.Add(-time.Minute),
		To:   base.Add(10 * time.Minute),
	})
	if err == nil {
		t.Fatal("expected an error when the range starts before the activation log")
	}
	if ch.UserStats["alice"].TotalActivations != before {
		t.Error("data was mutated even though the purge was refused")
	}

	// --partial only removes the events that actually exist in the log.
	result, err := PurgeUserActivations(ch, PurgeOptions{
		User:         "alice",
		From:         base.Add(-time.Minute),
		To:           base.Add(10 * time.Minute),
		AllowPartial: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Partial {
		t.Error("result should be marked partial")
	}
	if result.Removed != 1 {
		t.Errorf("Removed = %d, want 1", result.Removed)
	}
}

func TestPurgeUserActivationsWithoutLog(t *testing.T) {
	ch := purgeTestChannel()
	ch.ActivationLog = nil

	_, err := PurgeUserActivations(ch, PurgeOptions{
		User: "alice",
		From: time.Now().Add(-2 * time.Hour),
		To:   time.Now(),
	})
	if err != ErrNoActivationLog {
		t.Fatalf("err = %v, want ErrNoActivationLog", err)
	}
}

func TestRemoveUser(t *testing.T) {
	ch := purgeTestChannel()
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)

	removed, err := RemoveUser(ch, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if removed != 3 {
		t.Errorf("removed = %d, want 3 logged activations", removed)
	}
	if _, exists := ch.UserStats["alice"]; exists {
		t.Error("alice is still in the channel")
	}
	if _, exists := ch.UserStats["bob"]; !exists {
		t.Error("bob was removed too")
	}
	if got := ch.DailyActivations[bertDayKey(base)]; got != 1 {
		t.Errorf("channel daily = %d, want bob's 1", got)
	}
	if got := ch.HourlyActivations[bertHourKey(base)]; got != 1 {
		t.Errorf("channel 12:00 hour = %d, want bob's 1", got)
	}
	if _, exists := ch.HourlyActivations[bertHourKey(base.Add(time.Hour))]; exists {
		t.Error("alice's 13:00 hour should be gone")
	}
	if len(ch.ActivationLog) != 1 || ch.ActivationLog[0].Username != "bob" {
		t.Errorf("activation log = %+v, want only bob's activation", ch.ActivationLog)
	}
	if len(ch.GoldenzazabertEvents) != 1 || ch.GoldenzazabertEvents[0].Username != "bob" {
		t.Errorf("goldenzazabert events = %+v, want only bob's", ch.GoldenzazabertEvents)
	}
}

func TestAdminPurgePersistsAndBacksUp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stats.json")
	data := map[string]*ChannelData{"testchannel": purgeTestChannel()}
	st := store.New(path, &data)
	if err := st.EnsureDir(); err != nil {
		t.Fatal(err)
	}
	if err := st.Save(); err != nil {
		t.Fatal(err)
	}

	s := &BertService{store: st, Data: data}
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)
	result, err := s.PurgeUserActivations("testchannel", PurgeOptions{
		User: "alice",
		From: base.Add(20 * time.Minute),
		To:   base.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Removed != 2 {
		t.Fatalf("removed = %d, want 2", result.Removed)
	}

	backups, err := filepath.Glob(path + ".bak-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Fatalf("backups = %d, want 1", len(backups))
	}

	var reloaded map[string]*ChannelData
	reloadedStore := store.New(path, &reloaded)
	if _, err := reloadedStore.LoadIfExists(); err != nil {
		t.Fatal(err)
	}
	if got := reloaded["testchannel"].UserStats["alice"].TotalActivations; got != 1 {
		t.Errorf("saved alice total = %d, want 1", got)
	}
}
