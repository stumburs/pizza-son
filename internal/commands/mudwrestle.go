package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"strconv"
	"strings"
	"sync"
	"time"
)

type duelChallenge struct {
	challengerID   string
	challengerName string
	amount         int
	expiresAt      time.Time
}

var (
	pendingDuels   = make(map[string]*duelChallenge)
	pendingDuelsMu sync.Mutex
)

func init() {
	Register(bot.Command{
		Name:        "mudwrestle",
		Description: "Challenge another user to a mud wrestling match for pizza slices. The winner takes all.",
		Usage:       "!mudwrestle <user> <amount> | !mudwrestle accept | !mudwrestle stats [user]",
		Category:    bot.CategoryGames,
		Examples: []bot.CommandExample{
			{Input: "!mudwrestle @sweaty_man67 100", Output: "pizza_tm has challenged @sweaty_man67 to a mud wrestling match for 100 pizza slices! Type !mudwrestle accept within 60 seconds to accept."},
			{Input: "!mudwrestle accept", Output: "sweaty_man67 has won the wrestling duel against pizza_tm and takes 100 slices."},
			{Input: "!mudwrestle stats", Output: "pizza_tm mudwrestle stats - Wins: 69 | Losses: 0 | Slices won: 420 | Slices lost: 0"},
			{Input: "!mudwrestle stats @creamerman", Output: "creamerman mudwrestle stats - Wins: 4 | Losses: 7 | Slices won: 245 | Slices lost: 460"},
		},
		Handler: func(ctx bot.CommandContext) {
			// Exclude Discord
			if ctx.Message.Platform == models.PlatformDiscord {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "This is a Twitch exclusive command.")
				return
			}
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !mudwrestle <user> <amount> | !mudwrestle accept | !mudwrestle stats [user]")
				return
			}

			switch ctx.Args[0] {
			case "accept":
				handleAccept(ctx)
			case "stats":
				var targetID, targetName string
				if len(ctx.Args) > 1 {
					targetName = strings.ToLower(strings.TrimPrefix(ctx.Args[1], "@"))
					id, err := services.TwitchServiceInstance.GetUserID(targetName)
					if err != nil {
						ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Could not find user: "+targetName)
						return
					}
					targetID = id
				} else {
					targetID = ctx.Message.User.ID
					targetName = ctx.Message.User.DisplayName
				}

				st := services.MudwrestleServiceInstance.GetStats(targetID)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf(
					"%s mudwrestle stats - Wins: %d | Losses: %d | Slices won: %d | Slices lost: %d",
					targetName, st.Wins, st.Losses, st.SlicesWon, st.SlicesLost,
				))

			default:
				handleChallenge(ctx)
			}
		},
	})
}

func handleChallenge(ctx bot.CommandContext) {
	if len(ctx.Args) < 2 {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !mudwrestle <user> <amount>")
		return
	}

	target := strings.ToLower(strings.TrimPrefix(ctx.Args[0], "@"))

	if target == strings.ToLower(ctx.Message.User.Name) {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "You can't wrestle with yourself. unless...")
		return
	}

	amount, err := strconv.Atoi(ctx.Args[1])
	if err != nil || amount <= 0 {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Amount must be a positive number.")
		return
	}

	balance := services.CurrencyServiceInstance.Balance(ctx.Message.User.ID)
	if balance < amount {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("You don't have enough slices. You only have %d slices.", balance))
		return
	}

	targetID, err := services.TwitchServiceInstance.GetUserID(target)
	if err != nil {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Something went wrong when checking %ss balance.", target))
		return
	}

	targetBalance := services.CurrencyServiceInstance.Balance(targetID)
	if targetBalance < amount {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s doesn't have enough pizza slices to accept the duel.", target))
		return
	}

	pendingDuelsMu.Lock()
	pendingDuels[target] = &duelChallenge{
		challengerID:   ctx.Message.User.ID,
		challengerName: ctx.Message.User.DisplayName,
		amount:         amount,
		expiresAt:      time.Now().Add(60 * time.Second),
	}
	pendingDuelsMu.Unlock()

	ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("%s has challenged @%s to a mud wrestling match for %d pizza slices! Type !mudwrestle accept within 60 seconds to accept.", ctx.Message.User.DisplayName, target, amount))
}

func handleAccept(ctx bot.CommandContext) {
	accepterName := strings.ToLower(ctx.Message.User.Name)

	pendingDuelsMu.Lock()
	challenge, ok := pendingDuels[accepterName]
	if !ok {
		pendingDuelsMu.Unlock()
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "You have no pending challenges.")
		return
	}
	if time.Now().After(challenge.expiresAt) {
		delete(pendingDuels, accepterName)
		pendingDuelsMu.Unlock()
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "The mud wrestling challenge has expired.")
		return
	}
	delete(pendingDuels, accepterName)
	pendingDuelsMu.Unlock()

	// Check accepter has enough slices
	accepterBalance := services.CurrencyServiceInstance.Balance(ctx.Message.User.ID)
	if accepterBalance < challenge.amount {
		ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("@%s doesn't have enough pizza slices to accept challenge. They only have %d slices.", ctx.Message.User.DisplayName, accepterBalance))
		return
	}

	challengerWins := rand.IntN(2) == 0

	if challengerWins {
		services.CurrencyServiceInstance.Give(ctx.Message.User.ID, challenge.challengerID, challenge.amount)
		services.MudwrestleServiceInstance.RecordWin(challenge.challengerID, challenge.amount)
		services.MudwrestleServiceInstance.RecordLoss(ctx.Message.User.ID, challenge.amount)
		ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("%s has won the wrestling duel against @%s and takes %d slices.", challenge.challengerName, ctx.Message.User.DisplayName, challenge.amount))
	} else {
		services.CurrencyServiceInstance.Give(challenge.challengerID, ctx.Message.User.ID, challenge.amount)
		services.MudwrestleServiceInstance.RecordWin(ctx.Message.User.ID, challenge.amount)
		services.MudwrestleServiceInstance.RecordLoss(challenge.challengerID, challenge.amount)
		ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("%s has won the wrestling duel against @%s and takes %d slices.", ctx.Message.User.DisplayName, challenge.challengerName, challenge.amount))
	}
}
