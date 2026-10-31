package services

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"time"
)

// PurgeUserActivationsApprox removes an approximate number of one user's
// activations in a time range. It is meant for data that was recorded before
// the per-activation log existed: the number of activations to remove is
// derived from the user's hourly counters, and individual berts are then
// picked at random, weighted by how often the user rolled them.
//
// Any activations that *are* present in the log inside the range are removed
// exactly first. The remainder is approximated, so running this over a fully
// logged range still behaves like an exact purge.
func PurgeUserActivationsApprox(ch *ChannelData, opts PurgeOptions) (PurgeResult, error) {
	return purgeUserActivationsApprox(ch, opts, rand.IntN)
}

func purgeUserActivationsApprox(ch *ChannelData, opts PurgeOptions, intN func(int) int) (PurgeResult, error) {
	var result PurgeResult
	result.Approximate = true
	result.RemovedByBert = make(map[string]int)

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

	if stats.BertRecords == nil {
		stats.BertRecords = make(map[string]BertRecord)
	}
	if stats.DailyActivations == nil {
		stats.DailyActivations = make(map[string]int)
	}
	if stats.HourlyActivations == nil {
		stats.HourlyActivations = make(map[string]int)
	}

	result.LogStarts, result.LogEnds = activationLogRange(ch.ActivationLog)

	affectedBerts := make(map[string]struct{})

	// Remove anything that was actually logged inside the range exactly, so
	// the log stays consistent with the aggregates. When a total count was
	// requested, logged removals count towards it.
	var loggedGoldEvents []time.Time
	loggedRemoved := 0
	if len(ch.ActivationLog) > 0 {
		kept := make([]BertActivation, 0, len(ch.ActivationLog))
		for _, activation := range ch.ActivationLog {
			inRange := activation.Username == opts.User &&
				!activation.Timestamp.Before(opts.From) &&
				activation.Timestamp.Before(opts.To)
			if inRange && (opts.Count <= 0 || loggedRemoved < opts.Count) {
				applyActivationRemoval(ch, &stats, activation, &result)
				affectedBerts[activation.Bert] = struct{}{}
				loggedRemoved++
				if activation.Bert == "goldenzazabert" {
					loggedGoldEvents = append(loggedGoldEvents, activation.Timestamp)
				}
				continue
			}
			kept = append(kept, activation)
		}
		ch.ActivationLog = kept
	}

	// Logged goldenzazabert activations in the range took their events with
	// them. Only remove the matching events; unlogged events are handled by
	// the random selection below.
	if len(loggedGoldEvents) > 0 && len(ch.GoldenzazabertEvents) > 0 {
		goldTimes := make(map[int64]struct{}, len(loggedGoldEvents))
		for _, timestamp := range loggedGoldEvents {
			goldTimes[timestamp.UnixNano()] = struct{}{}
		}
		kept := make([]GoldenzazabertEvent, 0, len(ch.GoldenzazabertEvents))
		for _, event := range ch.GoldenzazabertEvents {
			if event.Username == opts.User {
				if _, remove := goldTimes[event.Timestamp.UnixNano()]; remove {
					continue
				}
			}
			kept = append(kept, event)
		}
		ch.GoldenzazabertEvents = kept
	}

	// Work out how many *unlogged* activations are left in the range. The
	// exact removals above already decremented the hourly counters, so the
	// remaining counters describe the part we have to approximate.
	hourPlan := approximateHourPlan(stats.HourlyActivations, opts.From, opts.To)
	planTotal := sumPlan(hourPlan)

	target := planTotal
	if opts.Count > 0 {
		target = opts.Count - loggedRemoved
		if target < 0 {
			target = 0
		}
	}
	if target <= 0 {
		finishApproxPurge(ch, &stats, opts, affectedBerts, &result)
		return result, nil
	}

	if stats.TotalActivations < target {
		target = stats.TotalActivations
		if opts.Count > 0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("only %d activation(s) were left to remove, fewer than requested", target))
		} else {
			result.Warnings = append(result.Warnings,
				"the hourly counters asked for more removals than the user has activations; the count was capped")
		}
	}

	eligible := eligibleBerts(stats.BertRecords, opts.From, opts.To)
	available := sumWeights(eligible)
	if available < target {
		// A record's First/LastSeen only describes the extremes, so the range
		// filter can be too strict on unusual data. Fall back to every record
		// rather than quietly removing fewer activations than requested.
		eligible = allBertWeights(stats.BertRecords)
		available = sumWeights(eligible)
		if available > 0 {
			result.Warnings = append(result.Warnings,
				"not enough bert records overlapped the range; random removal used all of the user's berts")
		}
	}
	if available == 0 {
		result.Warnings = append(result.Warnings,
			"no bert records are available to remove in this range")
		finishApproxPurge(ch, &stats, opts, affectedBerts, &result)
		return result, nil
	}
	if target > available {
		target = available
		result.Warnings = append(result.Warnings,
			"more activations were requested than the remaining bert records hold; the count was capped")
	}

	if opts.Count > 0 && target > planTotal {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"the hourly timeline only recorded %d activation(s) in this range; the bert records were still reduced by the requested amount",
			planTotal,
		))
	}

	approxRemoved := 0
	approxByBert := make(map[string]int)
	for approxRemoved < target {
		bert := pickWeightedBert(eligible, intN)
		if bert == "" {
			break
		}
		eligible[bert]--
		record := stats.BertRecords[bert]
		record.Count--
		stats.BertRecords[bert] = record
		approxRemoved++
		approxByBert[bert]++
		affectedBerts[bert] = struct{}{}
	}

	if approxRemoved == 0 {
		finishApproxPurge(ch, &stats, opts, affectedBerts, &result)
		return result, nil
	}

	// Decrement the timelines by the per-hour plan, trimmed to what was
	// actually removed.
	hourPlan = scaleHourPlan(hourPlan, approxRemoved)
	applyHourPlan(ch, &stats, hourPlan)

	stats.TotalActivations -= approxRemoved
	for bert, count := range approxByBert {
		result.Removed += count
		result.RemovedByBert[bert] += count
	}

	// Only drop goldenzazabert hall-of-fame events when the random selection
	// actually picked a goldenzazabert activation.
	if count := approxByBert["goldenzazabert"]; count > 0 {
		removedEvents := removeRandomGoldenzabertEvents(ch, opts.User, opts.From, opts.To, count, intN)
		if removedEvents < count {
			result.Warnings = append(result.Warnings, fmt.Sprintf(
				"only %d goldenzazabert event(s) could be removed, but %d activation(s) were selected",
				removedEvents, count,
			))
		}
	}

	clampRecordZazas(&stats, affectedBerts)
	refreshLastSeen(ch, &stats, opts.User, opts.From, opts.To, affectedBerts)
	clampUserZazaTotals(&stats, &result)
	finalizePurgedUser(ch, &stats, opts.User, &result)
	appendCountShortfallWarning(opts, &result)

	return result, nil
}

func finishApproxPurge(ch *ChannelData, stats *UserStats, opts PurgeOptions, affected map[string]struct{}, result *PurgeResult) {
	clampRecordZazas(stats, affected)
	refreshLastSeen(ch, stats, opts.User, opts.From, opts.To, affected)
	clampUserZazaTotals(stats, result)
	finalizePurgedUser(ch, stats, opts.User, result)
	appendCountShortfallWarning(opts, result)
}

func appendCountShortfallWarning(opts PurgeOptions, result *PurgeResult) {
	if opts.Count > 0 && result.Removed < opts.Count {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"only %d of the requested %d activation(s) could be removed",
			result.Removed, opts.Count,
		))
	}
}

// applyActivationRemoval subtracts one known activation from every aggregate.
func applyActivationRemoval(ch *ChannelData, stats *UserStats, activation BertActivation, result *PurgeResult) {
	result.Removed++
	result.RemovedByBert[activation.Bert]++

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

// approximateHourPlan distributes a user's hourly activity across the overlap
// with [from, to). Partial hours are prorated, then the total is rounded once
// and split across hours by largest remainder so many small partial hours do
// not each round up.
func approximateHourPlan(hourly map[string]int, from, to time.Time) map[string]int {
	type allocation struct {
		key      string
		expected float64
		units    int
		max      int
	}

	allocations := make([]allocation, 0, len(hourly))
	totalExpected := 0.0
	for key, count := range hourly {
		if count <= 0 {
			continue
		}
		start, err := time.ParseInLocation("2006-01-02 15:00", key, time.Local)
		if err != nil {
			continue
		}
		end := start.Add(time.Hour)

		overlapStart := start
		if from.After(overlapStart) {
			overlapStart = from
		}
		overlapEnd := end
		if to.Before(overlapEnd) {
			overlapEnd = to
		}
		if !overlapStart.Before(overlapEnd) {
			continue
		}

		fraction := overlapEnd.Sub(overlapStart).Minutes() / 60.0
		expected := float64(count) * fraction
		if expected > float64(count) {
			expected = float64(count)
		}
		if expected <= 0 {
			continue
		}

		allocations = append(allocations, allocation{
			key:      key,
			expected: expected,
			units:    int(math.Floor(expected)),
			max:      count,
		})
		totalExpected += expected
	}

	target := int(math.Round(totalExpected))
	assigned := 0
	for _, allocation := range allocations {
		assigned += allocation.units
	}

	sort.Slice(allocations, func(i, j int) bool {
		left := allocations[i].expected - math.Floor(allocations[i].expected)
		right := allocations[j].expected - math.Floor(allocations[j].expected)
		if left != right {
			return left > right
		}
		return allocations[i].key < allocations[j].key
	})

	// Hand out the remaining units to the hours with the largest fractions.
	for index := 0; assigned < target && index < len(allocations); {
		allocation := &allocations[index]
		if allocation.units < allocation.max {
			allocation.units++
			assigned++
		}
		index++
		if index == len(allocations) && assigned < target {
			index = 0
		}
	}

	plan := make(map[string]int)
	for _, allocation := range allocations {
		if allocation.units > 0 {
			plan[allocation.key] = allocation.units
		}
	}
	return plan
}

func sumPlan(plan map[string]int) int {
	total := 0
	for _, count := range plan {
		total += count
	}
	return total
}

// scaleHourPlan trims a plan so its total equals target. The exact
// distribution no longer matters once the data is approximate.
func scaleHourPlan(plan map[string]int, target int) map[string]int {
	if target <= 0 {
		return map[string]int{}
	}
	shortfall := sumPlan(plan) - target
	if shortfall <= 0 {
		return plan
	}

	keys := make([]string, 0, len(plan))
	for key := range plan {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if shortfall == 0 {
			break
		}
		take := plan[key]
		if take > shortfall {
			take = shortfall
		}
		plan[key] -= take
		shortfall -= take
		if plan[key] == 0 {
			delete(plan, key)
		}
	}
	return plan
}

// applyHourPlan decrements the channel and user timelines for the planned
// hours and returns the number of hours' worth of activations applied.
func applyHourPlan(ch *ChannelData, stats *UserStats, plan map[string]int) int {
	daily := make(map[string]int)
	total := 0

	for key, count := range plan {
		start, err := time.ParseInLocation("2006-01-02 15:00", key, time.Local)
		if err != nil {
			continue
		}
		decrementCounter(ch.HourlyActivations, key, count)
		decrementCounter(stats.HourlyActivations, key, count)
		daily[bertDayKey(start)] += count
		total += count
	}

	for day, count := range daily {
		decrementCounter(ch.DailyActivations, day, count)
		decrementCounter(stats.DailyActivations, day, count)
	}
	return total
}

func eligibleBerts(records map[string]BertRecord, from, to time.Time) map[string]int {
	weights := make(map[string]int)
	for bert, record := range records {
		if record.Count <= 0 {
			continue
		}
		if record.FirstSeen.IsZero() || record.LastSeen.IsZero() {
			weights[bert] = record.Count
			continue
		}
		if record.FirstSeen.Before(to) && !record.LastSeen.Before(from) {
			weights[bert] = record.Count
		}
	}
	return weights
}

func allBertWeights(records map[string]BertRecord) map[string]int {
	weights := make(map[string]int)
	for bert, record := range records {
		if record.Count > 0 {
			weights[bert] = record.Count
		}
	}
	return weights
}

func sumWeights(weights map[string]int) int {
	total := 0
	for _, count := range weights {
		total += count
	}
	return total
}

// pickWeightedBert chooses a bert with probability proportional to its
// remaining count. Keys are sorted so a deterministic intN makes tests stable.
func pickWeightedBert(weights map[string]int, intN func(int) int) string {
	keys := make([]string, 0, len(weights))
	total := 0
	for bert, count := range weights {
		if count > 0 {
			keys = append(keys, bert)
			total += count
		}
	}
	if total <= 0 {
		return ""
	}
	sort.Strings(keys)

	pick := intN(total)
	for _, bert := range keys {
		count := weights[bert]
		if pick < count {
			return bert
		}
		pick -= count
	}
	return keys[len(keys)-1]
}

// refreshLastSeen fixes records whose newest activation fell inside the
// purged range. If the log has nothing newer, FirstSeen is the best fallback.
func refreshLastSeen(ch *ChannelData, stats *UserStats, user string, from, to time.Time, berts map[string]struct{}) {
	for bert := range berts {
		record, exists := stats.BertRecords[bert]
		if !exists {
			continue
		}
		if record.Count <= 0 {
			delete(stats.BertRecords, bert)
			continue
		}
		if record.LastSeen.IsZero() || (!record.LastSeen.Before(from) && record.LastSeen.Before(to)) {
			var latest time.Time
			for _, activation := range ch.ActivationLog {
				if activation.Username == user && activation.Bert == bert && activation.Timestamp.After(latest) {
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
}

// clampRecordZazas keeps per-bert zaza counters from exceeding the number of
// activations that are left. Random removals cannot know which activation was
// a zaza, so clamping is the honest option.
func clampRecordZazas(stats *UserStats, berts map[string]struct{}) {
	for bert := range berts {
		record, exists := stats.BertRecords[bert]
		if !exists {
			continue
		}
		if record.Count <= 0 {
			stats.TotalZazas -= record.ZazaCount
			stats.TotalZazaLs -= record.ZazaLCount
			stats.TotalDoubleZazas -= record.DoubleZazaCount
			delete(stats.BertRecords, bert)
			continue
		}
		if record.ZazaCount > record.Count {
			stats.TotalZazas -= record.ZazaCount - record.Count
			record.ZazaCount = record.Count
		}
		if record.ZazaLCount > record.Count {
			stats.TotalZazaLs -= record.ZazaLCount - record.Count
			record.ZazaLCount = record.Count
		}
		if record.DoubleZazaCount > record.Count {
			stats.TotalDoubleZazas -= record.DoubleZazaCount - record.Count
			record.DoubleZazaCount = record.Count
		}
		stats.BertRecords[bert] = record
	}
}

func clampUserZazaTotals(stats *UserStats, result *PurgeResult) {
	stats.TotalZazas = clampCounter(&stats.TotalZazas, "total zazas", &result.Warnings)
	stats.TotalZazaLs = clampCounter(&stats.TotalZazaLs, "total zazaLs", &result.Warnings)
	stats.TotalDoubleZazas = clampCounter(&stats.TotalDoubleZazas, "total double zazas", &result.Warnings)
}

func finalizePurgedUser(ch *ChannelData, stats *UserStats, user string, result *PurgeResult) {
	stats.TotalActivations = clampCounter(&stats.TotalActivations, "total activations", &result.Warnings)

	if stats.TotalActivations == 0 && len(stats.BertRecords) == 0 {
		delete(ch.UserStats, user)
		result.UserRemoved = true
		return
	}
	ch.UserStats[user] = *stats
}

// removeRandomGoldenzabertEvents drops count random goldenzazabert events for
// the user inside the range.
func removeRandomGoldenzabertEvents(ch *ChannelData, user string, from, to time.Time, count int, intN func(int) int) int {
	if count <= 0 || len(ch.GoldenzazabertEvents) == 0 {
		return 0
	}

	indices := make([]int, 0, len(ch.GoldenzazabertEvents))
	for i, event := range ch.GoldenzazabertEvents {
		if event.Username == user && !event.Timestamp.Before(from) && event.Timestamp.Before(to) {
			indices = append(indices, i)
		}
	}
	if len(indices) == 0 {
		return 0
	}
	if count > len(indices) {
		count = len(indices)
	}

	drop := make(map[int]struct{}, count)
	for picked := 0; picked < count; picked++ {
		pos := intN(len(indices))
		drop[indices[pos]] = struct{}{}
		indices[pos] = indices[len(indices)-1]
		indices = indices[:len(indices)-1]
	}

	kept := make([]GoldenzazabertEvent, 0, len(ch.GoldenzazabertEvents)-count)
	for i, event := range ch.GoldenzazabertEvents {
		if _, remove := drop[i]; remove {
			continue
		}
		kept = append(kept, event)
	}
	ch.GoldenzazabertEvents = kept
	return count
}
