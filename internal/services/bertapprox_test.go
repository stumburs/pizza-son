package services

import (
	"path/filepath"
	"pizza-son/internal/store"
	"strings"
	"testing"
	"time"
)

func approxBase() time.Time {
	return time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)
}

// approxTestChannel has no activation log, which is the case this mode was
// built for: older aggregate-only data.
func approxTestChannel() *ChannelData {
	base := approxBase()
	return &ChannelData{
		Berts: []string{"bertnard", "beert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 4,
				TotalZazas:       2,
				BertRecords: map[string]BertRecord{
					"bertnard": {
						Count:     2,
						ZazaCount: 1,
						FirstSeen: base,
						LastSeen:  base.Add(90 * time.Minute),
					},
					"beert": {
						Count:     2,
						ZazaCount: 1,
						FirstSeen: base.Add(10 * time.Minute),
						LastSeen:  base.Add(70 * time.Minute),
					},
				},
				DailyActivations: map[string]int{bertDayKey(base): 4},
				HourlyActivations: map[string]int{
					bertHourKey(base):                2,
					bertHourKey(base.Add(time.Hour)): 2,
				},
			},
			"bob": {
				TotalActivations: 1,
				BertRecords: map[string]BertRecord{
					"beert": {
						Count:     1,
						FirstSeen: base.Add(70 * time.Minute),
						LastSeen:  base.Add(70 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 1},
				HourlyActivations: map[string]int{bertHourKey(base.Add(time.Hour)): 1},
			},
		},
		DailyActivations: map[string]int{bertDayKey(base): 5},
		HourlyActivations: map[string]int{
			bertHourKey(base):                2,
			bertHourKey(base.Add(time.Hour)): 3,
		},
		GoldenzazabertEvents: []GoldenzazabertEvent{
			{Username: "alice", Timestamp: base.Add(10 * time.Minute)},
			{Username: "alice", Timestamp: base.Add(70 * time.Minute)},
			{Username: "bob", Timestamp: base.Add(75 * time.Minute)},
		},
	}
}

func TestApproxPurgeWithoutLogUsesHourlyCounts(t *testing.T) {
	ch := approxTestChannel()
	base := approxBase()

	result, err := purgeUserActivationsApprox(ch, PurgeOptions{
		User: "alice",
		From: base,
		To:   base.Add(time.Hour),
	}, func(int) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}

	if !result.Approximate {
		t.Error("result should be marked approximate")
	}
	if result.Removed != 2 {
		t.Fatalf("removed = %d, want 2", result.Removed)
	}
	if result.RemovedByBert["beert"] != 2 {
		t.Errorf("removed_by_bert = %v, want beert ×2", result.RemovedByBert)
	}

	alice := ch.UserStats["alice"]
	if alice.TotalActivations != 2 {
		t.Errorf("alice total = %d, want 2", alice.TotalActivations)
	}
	if alice.TotalZazas != 1 {
		t.Errorf("alice zazas = %d, want 1 after clamping the deleted beert record", alice.TotalZazas)
	}
	if _, exists := alice.BertRecords["beert"]; exists {
		t.Error("beert record should be gone")
	}
	bertnard := alice.BertRecords["bertnard"]
	if bertnard.Count != 2 || bertnard.ZazaCount != 1 {
		t.Errorf("bertnard record = %+v, want count 2 with 1 zaza", bertnard)
	}
	if _, exists := alice.HourlyActivations[bertHourKey(base)]; exists {
		t.Error("alice's first hour should be empty")
	}
	if got := alice.HourlyActivations[bertHourKey(base.Add(time.Hour))]; got != 2 {
		t.Errorf("alice's second hour = %d, want 2", got)
	}
	if got := alice.DailyActivations[bertDayKey(base)]; got != 2 {
		t.Errorf("alice's daily = %d, want 2", got)
	}

	if _, exists := ch.HourlyActivations[bertHourKey(base)]; exists {
		t.Error("channel's first hour should be empty")
	}
	if got := ch.HourlyActivations[bertHourKey(base.Add(time.Hour))]; got != 3 {
		t.Errorf("channel's second hour = %d, want 3", got)
	}
	if got := ch.DailyActivations[bertDayKey(base)]; got != 3 {
		t.Errorf("channel daily = %d, want 3", got)
	}

	if len(ch.GoldenzazabertEvents) != 3 {
		t.Errorf("goldenzazabert events = %+v, want the original 3", ch.GoldenzazabertEvents)
	}
}

func TestApproxPurgeRemovesLoggedActivationsExactlyFirst(t *testing.T) {
	base := approxBase()
	ch := &ChannelData{
		Berts: []string{"bertnard", "beert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 2,
				TotalZazas:       1,
				BertRecords: map[string]BertRecord{
					"beert": {
						Count:     2,
						ZazaCount: 1,
						FirstSeen: base.Add(10 * time.Minute),
						LastSeen:  base.Add(10 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 2},
				HourlyActivations: map[string]int{bertHourKey(base): 2},
			},
			"bob": {
				TotalActivations: 1,
				BertRecords: map[string]BertRecord{
					"beert": {
						Count:     1,
						FirstSeen: base.Add(5 * time.Minute),
						LastSeen:  base.Add(5 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 1},
				HourlyActivations: map[string]int{bertHourKey(base): 1},
			},
		},
		DailyActivations:  map[string]int{bertDayKey(base): 3},
		HourlyActivations: map[string]int{bertHourKey(base): 3},
		ActivationLog: []BertActivation{
			{Username: "alice", Bert: "beert", Timestamp: base.Add(10 * time.Minute), IsZaza: true},
			{Username: "bob", Bert: "beert", Timestamp: base.Add(5 * time.Minute)},
		},
	}

	result, err := purgeUserActivationsApprox(ch, PurgeOptions{
		User: "alice",
		From: base,
		To:   base.Add(time.Hour),
	}, func(int) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}

	if result.Removed != 2 {
		t.Fatalf("removed = %d, want 2 (one logged + one estimated)", result.Removed)
	}
	if result.RemovedByBert["beert"] != 2 {
		t.Errorf("removed_by_bert = %v, want beert ×2", result.RemovedByBert)
	}
	if len(ch.ActivationLog) != 1 || ch.ActivationLog[0].Username != "bob" {
		t.Errorf("log = %+v, want only bob's activation", ch.ActivationLog)
	}

	if !result.UserRemoved {
		t.Error("alice should be removed after both of her activations were purged")
	}
	if _, exists := ch.UserStats["alice"]; exists {
		t.Error("alice is still in the channel")
	}
	if _, exists := ch.UserStats["bob"]; !exists {
		t.Error("bob should still exist")
	}
	if got := ch.HourlyActivations[bertHourKey(base)]; got != 1 {
		t.Errorf("channel hour = %d, want bob's 1", got)
	}
	if got := ch.DailyActivations[bertDayKey(base)]; got != 1 {
		t.Errorf("channel daily = %d, want bob's 1", got)
	}
}

func TestApproxPurgePicksGoldenzazabertEvents(t *testing.T) {
	base := approxBase()
	ch := &ChannelData{
		Berts: []string{"goldenzazabert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 2,
				BertRecords: map[string]BertRecord{
					"goldenzazabert": {
						Count:     2,
						FirstSeen: base,
						LastSeen:  base.Add(40 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 2},
				HourlyActivations: map[string]int{bertHourKey(base): 2},
			},
		},
		DailyActivations:  map[string]int{bertDayKey(base): 2},
		HourlyActivations: map[string]int{bertHourKey(base): 2},
		GoldenzazabertEvents: []GoldenzazabertEvent{
			{Username: "alice", Timestamp: base.Add(10 * time.Minute)},
			{Username: "alice", Timestamp: base.Add(20 * time.Minute)},
			{Username: "alice", Timestamp: base.Add(30 * time.Minute)},
			{Username: "alice", Timestamp: base.Add(2 * time.Hour)},
			{Username: "bob", Timestamp: base.Add(10 * time.Minute)},
		},
	}

	result, err := purgeUserActivationsApprox(ch, PurgeOptions{
		User: "alice",
		From: base,
		To:   base.Add(time.Hour),
	}, func(int) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}

	if result.RemovedByBert["goldenzazabert"] != 2 {
		t.Errorf("removed_by_bert = %v, want goldenzazabert ×2", result.RemovedByBert)
	}
	if !result.UserRemoved {
		t.Error("alice should be removed after her only record was purged")
	}

	// intN 0 picks the first two in-range alice events; the third in-range
	// event, the out-of-range event and bob's event survive.
	if len(ch.GoldenzazabertEvents) != 3 {
		t.Fatalf("events = %+v, want 3 left", ch.GoldenzazabertEvents)
	}
	var aliceEvents []GoldenzazabertEvent
	for _, event := range ch.GoldenzazabertEvents {
		if event.Username == "alice" {
			aliceEvents = append(aliceEvents, event)
		}
	}
	if len(aliceEvents) != 2 {
		t.Fatalf("alice events = %+v, want 2 left", aliceEvents)
	}
	if !aliceEvents[1].Timestamp.Equal(base.Add(2 * time.Hour)) {
		t.Errorf("out-of-range alice event should survive, got %+v", aliceEvents)
	}
	inRange := 0
	for _, event := range aliceEvents {
		if !event.Timestamp.Before(base) && event.Timestamp.Before(base.Add(time.Hour)) {
			inRange++
		}
	}
	if inRange != 1 {
		t.Errorf("in-range alice events left = %d, want 1 (two randomly removed)", inRange)
	}
}

func TestApproxPurgeProratesPartialHours(t *testing.T) {
	base := approxBase()
	ch := &ChannelData{
		Berts: []string{"beert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 4,
				BertRecords: map[string]BertRecord{
					"beert": {
						Count:     4,
						FirstSeen: base,
						LastSeen:  base.Add(50 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 4},
				HourlyActivations: map[string]int{bertHourKey(base): 4},
			},
		},
		DailyActivations:  map[string]int{bertDayKey(base): 4},
		HourlyActivations: map[string]int{bertHourKey(base): 4},
	}

	result, err := purgeUserActivationsApprox(ch, PurgeOptions{
		User: "alice",
		From: base.Add(30 * time.Minute),
		To:   base.Add(time.Hour),
	}, func(int) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}

	if result.Removed != 2 {
		t.Fatalf("removed = %d, want half of the hour's 4 activations", result.Removed)
	}
	alice := ch.UserStats["alice"]
	if alice.TotalActivations != 2 || alice.BertRecords["beert"].Count != 2 {
		t.Errorf("alice total/record = %d/%d, want 2/2",
			alice.TotalActivations, alice.BertRecords["beert"].Count)
	}
	if got := ch.HourlyActivations[bertHourKey(base)]; got != 2 {
		t.Errorf("channel hour = %d, want 2", got)
	}
	if got := alice.HourlyActivations[bertHourKey(base)]; got != 2 {
		t.Errorf("alice hour = %d, want 2", got)
	}
}

func TestApproxPurgeFallsBackWhenRecordsDoNotOverlap(t *testing.T) {
	base := approxBase()
	ch := &ChannelData{
		Berts: []string{"beert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 1,
				BertRecords: map[string]BertRecord{
					"beert": {
						Count:     1,
						FirstSeen: base.Add(-10 * time.Hour),
						LastSeen:  base.Add(-10 * time.Hour),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 1},
				HourlyActivations: map[string]int{bertHourKey(base): 1},
			},
		},
		DailyActivations:  map[string]int{bertDayKey(base): 1},
		HourlyActivations: map[string]int{bertHourKey(base): 1},
	}

	result, err := purgeUserActivationsApprox(ch, PurgeOptions{
		User: "alice",
		From: base,
		To:   base.Add(time.Hour),
	}, func(int) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}

	if result.Removed != 1 {
		t.Errorf("removed = %d, want 1", result.Removed)
	}
	if result.UserRemoved != true {
		t.Error("alice should be removed")
	}
	found := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "not enough bert records overlapped") {
			found = true
		}
	}
	if !found {
		t.Errorf("warnings = %v, want a fallback warning", result.Warnings)
	}
}

func TestApproxPurgeRemovesOnlyMatchingLoggedGoldEvents(t *testing.T) {
	base := approxBase()
	ch := &ChannelData{
		Berts: []string{"goldenzazabert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 3,
				BertRecords: map[string]BertRecord{
					"goldenzazabert": {
						Count:     3,
						FirstSeen: base,
						LastSeen:  base.Add(30 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 3},
				HourlyActivations: map[string]int{bertHourKey(base): 3},
			},
		},
		DailyActivations:  map[string]int{bertDayKey(base): 3},
		HourlyActivations: map[string]int{bertHourKey(base): 3},
		GoldenzazabertEvents: []GoldenzazabertEvent{
			{Username: "alice", Timestamp: base.Add(10 * time.Minute)},
			{Username: "alice", Timestamp: base.Add(20 * time.Minute)},
			{Username: "alice", Timestamp: base.Add(30 * time.Minute)},
		},
		ActivationLog: []BertActivation{
			{Username: "alice", Bert: "goldenzazabert", Timestamp: base.Add(10 * time.Minute)},
		},
	}

	result, err := purgeUserActivationsApprox(ch, PurgeOptions{
		User: "alice",
		From: base,
		To:   base.Add(30 * time.Minute),
	}, func(int) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}

	if result.Removed != 2 {
		t.Fatalf("removed = %d, want 2 (one logged + one estimated)", result.Removed)
	}
	if len(ch.GoldenzazabertEvents) != 1 {
		t.Fatalf("events = %+v, want exactly one left", ch.GoldenzazabertEvents)
	}
	if !ch.GoldenzazabertEvents[0].Timestamp.Equal(base.Add(30 * time.Minute)) {
		t.Errorf("remaining event = %+v, want the unlogged 12:30 event", ch.GoldenzazabertEvents[0])
	}
}

func TestApproxPurgeHonorsExplicitCount(t *testing.T) {
	base := approxBase()
	ch := &ChannelData{
		Berts: []string{"beert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 10,
				BertRecords: map[string]BertRecord{
					"beert": {
						Count:     10,
						FirstSeen: base,
						LastSeen:  base.Add(50 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 10},
				HourlyActivations: map[string]int{bertHourKey(base): 10},
			},
		},
		DailyActivations:  map[string]int{bertDayKey(base): 10},
		HourlyActivations: map[string]int{bertHourKey(base): 10},
	}

	result, err := purgeUserActivationsApprox(ch, PurgeOptions{
		User:  "alice",
		From:  base,
		To:    base.Add(time.Hour),
		Count: 3,
	}, func(int) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}

	if result.Removed != 3 {
		t.Fatalf("removed = %d, want exactly 3", result.Removed)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", result.Warnings)
	}
	alice := ch.UserStats["alice"]
	if alice.TotalActivations != 7 || alice.BertRecords["beert"].Count != 7 {
		t.Errorf("alice total/record = %d/%d, want 7/7",
			alice.TotalActivations, alice.BertRecords["beert"].Count)
	}
	if got := ch.HourlyActivations[bertHourKey(base)]; got != 7 {
		t.Errorf("channel hour = %d, want 7", got)
	}
	if got := ch.DailyActivations[bertDayKey(base)]; got != 7 {
		t.Errorf("channel daily = %d, want 7", got)
	}
}

func TestApproxPurgeCountIncludesLoggedRemovals(t *testing.T) {
	base := approxBase()
	ch := &ChannelData{
		Berts: []string{"beert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 4,
				BertRecords: map[string]BertRecord{
					"beert": {
						Count:     4,
						FirstSeen: base,
						LastSeen:  base.Add(50 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 4},
				HourlyActivations: map[string]int{bertHourKey(base): 4},
			},
		},
		DailyActivations:  map[string]int{bertDayKey(base): 4},
		HourlyActivations: map[string]int{bertHourKey(base): 4},
		ActivationLog: []BertActivation{
			{Username: "alice", Bert: "beert", Timestamp: base.Add(10 * time.Minute)},
			{Username: "alice", Bert: "beert", Timestamp: base.Add(20 * time.Minute)},
		},
	}

	result, err := purgeUserActivationsApprox(ch, PurgeOptions{
		User:  "alice",
		From:  base,
		To:    base.Add(time.Hour),
		Count: 3,
	}, func(int) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}

	if result.Removed != 3 || result.RemovedByBert["beert"] != 3 {
		t.Errorf("result = %+v, want exactly 3 beert removals", result)
	}
	if len(ch.ActivationLog) != 0 {
		t.Errorf("log = %+v, want the two logged activations removed", ch.ActivationLog)
	}
	alice := ch.UserStats["alice"]
	if alice.TotalActivations != 1 || alice.BertRecords["beert"].Count != 1 {
		t.Errorf("alice total/record = %d/%d, want 1/1",
			alice.TotalActivations, alice.BertRecords["beert"].Count)
	}
}

func TestApproxPurgeWarnsWhenCountExceedsAvailable(t *testing.T) {
	base := approxBase()
	ch := &ChannelData{
		Berts: []string{"beert"},
		UserStats: map[string]UserStats{
			"alice": {
				TotalActivations: 2,
				BertRecords: map[string]BertRecord{
					"beert": {
						Count:     2,
						FirstSeen: base,
						LastSeen:  base.Add(50 * time.Minute),
					},
				},
				DailyActivations:  map[string]int{bertDayKey(base): 2},
				HourlyActivations: map[string]int{bertHourKey(base): 2},
			},
		},
		DailyActivations:  map[string]int{bertDayKey(base): 2},
		HourlyActivations: map[string]int{bertHourKey(base): 2},
	}

	result, err := purgeUserActivationsApprox(ch, PurgeOptions{
		User:  "alice",
		From:  base,
		To:    base.Add(time.Hour),
		Count: 5,
	}, func(int) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}

	if result.Removed != 2 {
		t.Errorf("removed = %d, want the 2 available activations", result.Removed)
	}
	found := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "of the requested 5") {
			found = true
		}
	}
	if !found {
		t.Errorf("warnings = %v, want a shortfall warning", result.Warnings)
	}
}

func TestAdminApproxPurgePersistsAndBacksUp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stats.json")
	data := map[string]*ChannelData{"testchannel": approxTestChannel()}
	st := store.New(path, &data)
	if err := st.EnsureDir(); err != nil {
		t.Fatal(err)
	}
	if err := st.Save(); err != nil {
		t.Fatal(err)
	}

	s := &BertService{store: st, Data: data}
	base := approxBase()
	result, err := s.PurgeUserActivationsApprox("testchannel", PurgeOptions{
		User: "alice",
		From: base,
		To:   base.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Removed != 4 || !result.Approximate {
		t.Fatalf("result = %+v, want 4 approximate removals", result)
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
	if _, exists := reloaded["testchannel"].UserStats["alice"]; exists {
		t.Error("alice should be gone from the saved stats")
	}
	if _, exists := reloaded["testchannel"].UserStats["bob"]; !exists {
		t.Error("bob should still be saved")
	}
}
