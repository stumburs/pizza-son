package services

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

// ErrNoActivationLog is returned when a channel only has aggregate data,
// i.e. it was recorded before per-activation logging was introduced.
var ErrNoActivationLog = errors.New("this channel has no per-activation log yet; use approximate purge for older data or wait for new bertchecks to be logged")

// PurgeOptions describes a time-range deletion of one user's activations.
type PurgeOptions struct {
	User string
	From time.Time // inclusive
	To   time.Time // exclusive

	// AllowPartial removes the logged activations inside the range even when
	// the log does not reach back to From. Aggregates recorded before the log
	// started are left untouched.
	AllowPartial bool

	// Count, when greater than zero, is the total number of activations to
	// remove in approximate mode instead of estimating the count from the
	// hourly activity. It is ignored by the exact purge.
	Count int
}

// PurgeResult reports what PurgeUserActivations changed.
type PurgeResult struct {
	Removed       int
	RemovedByBert map[string]int
	LogStarts     time.Time
	LogEnds       time.Time
	Partial       bool
	Approximate   bool
	UserRemoved   bool
	Warnings      []string
}

// PurgeUserActivations removes one user's activations in [From, To) from a
// channel and decrements every aggregate that was derived from them. It is
// exact as long as the activation log covers the requested range; the
// AllowPartial option disables that safety check.
func PurgeUserActivations(ch *ChannelData, opts PurgeOptions) (PurgeResult, error) {
	var result PurgeResult

	if ch == nil {
		return result, errors.New("channel not found")
	}
	if opts.User == "" {
		return result, errors.New("no user specified")
	}
	if !opts.From.Before(opts.To) {
		return result, errors.New("the start of the range must be before the end")
	}

	stats, ok := ch.UserStats[opts.User]
	if !ok {
		return result, fmt.Errorf("user %q has no stats in this channel", opts.User)
	}

	if len(ch.ActivationLog) == 0 {
		result.LogStarts, result.LogEnds = activationLogRange(ch.ActivationLog)
		return result, ErrNoActivationLog
	}

	result.RemovedByBert = make(map[string]int)
	result.LogStarts, result.LogEnds = activationLogRange(ch.ActivationLog)
	result.Partial = opts.From.Before(result.LogStarts)
	if result.Partial && !opts.AllowPartial {
		return result, fmt.Errorf(
			"activation log starts at %s, but the requested range starts at %s; enable partial removal to delete only the logged activations, or use approximate purge for the older data",
			result.LogStarts.Local().Format(time.RFC3339),
			opts.From.Local().Format(time.RFC3339),
		)
	}

	if stats.BertRecords == nil {
		stats.BertRecords = make(map[string]BertRecord)
	}
	if stats.DailyActivations == nil {
		stats.DailyActivations = make(map[string]int)
	}
	if stats.HourlyActivations == nil {
		stats.HourlyActivations = make(map[string]int)
	}

	affectedBerts := make(map[string]struct{})
	remaining := make([]BertActivation, 0, len(ch.ActivationLog))

	for _, activation := range ch.ActivationLog {
		if activation.Username != opts.User ||
			activation.Timestamp.Before(opts.From) ||
			!activation.Timestamp.Before(opts.To) {
			remaining = append(remaining, activation)
			continue
		}

		result.Removed++
		result.RemovedByBert[activation.Bert]++
		affectedBerts[activation.Bert] = struct{}{}

		decrementCounter(ch.DailyActivations, bertDayKey(activation.Timestamp), 1)
		decrementCounter(ch.HourlyActivations, bertHourKey(activation.Timestamp), 1)
		decrementCounter(stats.DailyActivations, bertDayKey(activation.Timestamp), 1)
		decrementCounter(stats.HourlyActivations, bertHourKey(activation.Timestamp), 1)

		stats.TotalActivations--
		if activation.IsZaza {
			stats.TotalZazas--
		}
		if activation.IsZazaL {
			stats.TotalZazaLs--
		}
		if activation.DoubleZaza {
			stats.TotalDoubleZazas--
		}

		record, exists := stats.BertRecords[activation.Bert]
		if !exists {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("activation for %q had no bert record; counts may be inconsistent", activation.Bert))
		}
		record.Count--
		if activation.IsZaza {
			record.ZazaCount--
		}
		if activation.IsZazaL {
			record.ZazaLCount--
		}
		if activation.DoubleZaza {
			record.DoubleZazaCount--
		}
		stats.BertRecords[activation.Bert] = record
	}

	if result.Removed == 0 {
		return result, nil
	}

	ch.ActivationLog = remaining

	// Drop goldenzazabert hall-of-fame entries that belong to removed events.
	if len(ch.GoldenzazabertEvents) > 0 {
		kept := make([]GoldenzazabertEvent, 0, len(ch.GoldenzazabertEvents))
		for _, event := range ch.GoldenzazabertEvents {
			if event.Username == opts.User &&
				!event.Timestamp.Before(opts.From) &&
				event.Timestamp.Before(opts.To) {
				continue
			}
			kept = append(kept, event)
		}
		ch.GoldenzazabertEvents = kept
	}

	// Recompute LastSeen for bert records whose most recent activation was
	// deleted. Older activations may predate the log, so fall back to
	// FirstSeen when nothing is left to point at.
	for bert := range affectedBerts {
		record, exists := stats.BertRecords[bert]
		if !exists {
			continue
		}
		if record.Count <= 0 {
			delete(stats.BertRecords, bert)
			continue
		}
		if record.LastSeen.IsZero() || (!record.LastSeen.Before(opts.From) && record.LastSeen.Before(opts.To)) {
			var latest time.Time
			for _, activation := range ch.ActivationLog {
				if activation.Username == opts.User && activation.Bert == bert && activation.Timestamp.After(latest) {
					latest = activation.Timestamp
				}
			}
			if latest.IsZero() {
				latest = record.FirstSeen
			}
			record.LastSeen = latest
			stats.BertRecords[bert] = record
		}
	}

	// A deleted activation should never leave a negative aggregate behind,
	// but old hand-edited or partially migrated files can be inconsistent.
	stats.TotalActivations = clampCounter(&stats.TotalActivations, "total activations", &result.Warnings)
	stats.TotalZazas = clampCounter(&stats.TotalZazas, "total zazas", &result.Warnings)
	stats.TotalZazaLs = clampCounter(&stats.TotalZazaLs, "total zazaLs", &result.Warnings)
	stats.TotalDoubleZazas = clampCounter(&stats.TotalDoubleZazas, "total double zazas", &result.Warnings)

	if stats.TotalActivations == 0 && len(stats.BertRecords) == 0 {
		delete(ch.UserStats, opts.User)
		result.UserRemoved = true
	} else {
		ch.UserStats[opts.User] = stats
	}

	return result, nil
}

// RemoveUser deletes a user from a channel and subtracts their recorded
// activity from the channel timeline. All of the user's logged activations
// are removed as well.
func RemoveUser(ch *ChannelData, user string) (int, error) {
	if ch == nil {
		return 0, errors.New("channel not found")
	}
	if user == "" {
		return 0, errors.New("no user specified")
	}

	stats, ok := ch.UserStats[user]
	if !ok {
		return 0, fmt.Errorf("user %q has no stats in this channel", user)
	}

	for day, count := range stats.DailyActivations {
		decrementCounter(ch.DailyActivations, day, count)
	}
	for hour, count := range stats.HourlyActivations {
		decrementCounter(ch.HourlyActivations, hour, count)
	}

	removed := 0
	if len(ch.ActivationLog) > 0 {
		kept := make([]BertActivation, 0, len(ch.ActivationLog))
		for _, activation := range ch.ActivationLog {
			if activation.Username == user {
				removed++
				continue
			}
			kept = append(kept, activation)
		}
		ch.ActivationLog = kept
	}

	if len(ch.GoldenzazabertEvents) > 0 {
		kept := make([]GoldenzazabertEvent, 0, len(ch.GoldenzazabertEvents))
		for _, event := range ch.GoldenzazabertEvents {
			if event.Username == user {
				continue
			}
			kept = append(kept, event)
		}
		ch.GoldenzazabertEvents = kept
	}

	delete(ch.UserStats, user)
	return removed, nil
}

func activationLogRange(log []BertActivation) (time.Time, time.Time) {
	var starts, ends time.Time
	for _, activation := range log {
		if starts.IsZero() || activation.Timestamp.Before(starts) {
			starts = activation.Timestamp
		}
		if ends.IsZero() || activation.Timestamp.After(ends) {
			ends = activation.Timestamp
		}
	}
	return starts, ends
}

func bertDayKey(t time.Time) string {
	return t.Format("2006-01-02")
}

func bertHourKey(t time.Time) string {
	return t.Format("2006-01-02 15:00")
}

func decrementCounter(m map[string]int, key string, by int) {
	if m == nil || by == 0 {
		return
	}
	next := m[key] - by
	if next <= 0 {
		delete(m, key)
		return
	}
	m[key] = next
}

func clampCounter(value *int, label string, warnings *[]string) int {
	if *value < 0 {
		*warnings = append(*warnings,
			fmt.Sprintf("%s went negative; clamped to 0 (the file was already inconsistent)", label))
		return 0
	}
	return *value
}

// PurgeUserActivations removes a user's activations from the live service and
// persists the result.
func (s *BertService) PurgeUserActivations(channel string, opts PurgeOptions) (PurgeResult, error) {
	return s.adminPurge(channel, "purge", func(ch *ChannelData) (PurgeResult, error) {
		return PurgeUserActivations(ch, opts)
	})
}

// PurgeUserActivationsApprox removes an estimated number of a user's
// activations from the live service and persists the result.
func (s *BertService) PurgeUserActivationsApprox(channel string, opts PurgeOptions) (PurgeResult, error) {
	return s.adminPurge(channel, "approximate purge", func(ch *ChannelData) (PurgeResult, error) {
		return PurgeUserActivationsApprox(ch, opts)
	})
}

// adminPurge locks the service, backs up the stats file and runs a purge
// function. The save only happens when something was removed.
func (s *BertService) adminPurge(channel, action string, purge func(*ChannelData) (PurgeResult, error)) (PurgeResult, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	ch, ok := s.Data[channel]
	if !ok {
		return PurgeResult{}, fmt.Errorf("channel %q not found", channel)
	}
	if err := s.backupForAdminChange(); err != nil {
		return PurgeResult{}, fmt.Errorf("backup stats before %s: %w", action, err)
	}

	result, err := purge(ch)
	if err != nil {
		return result, err
	}
	if result.Removed > 0 {
		s.save()
	}
	return result, nil
}

// RemoveUser deletes a user from a channel in the live service and persists
// the result.
func (s *BertService) RemoveUser(channel, user string) (int, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	ch, ok := s.Data[channel]
	if !ok {
		return 0, fmt.Errorf("channel %q not found", channel)
	}
	if err := s.backupForAdminChange(); err != nil {
		return 0, fmt.Errorf("backup stats before deleting user: %w", err)
	}

	removed, err := RemoveUser(ch, user)
	if err != nil {
		return 0, err
	}
	s.save()
	return removed, nil
}

// backupForAdminChange copies the current stats file next to itself so a
// destructive panel action can be undone manually. A missing file is fine:
// there is nothing to back up yet.
func (s *BertService) backupForAdminChange() error {
	if s.store == nil {
		return nil
	}

	path := s.store.Path()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	backupPath := fmt.Sprintf("%s.bak-%s", path, time.Now().Format("20060102-150405.000000000"))
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return err
	}
	log.Printf("[Bert] Admin backup written to %s", backupPath)
	return nil
}
