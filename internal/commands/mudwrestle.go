package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
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
		Usage:       "!mudwrestle <user> <amount> | !mudwrestle accept",
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !mudwrestle <user> <amount> | !mudwrestle accept")
				return
			}

			switch ctx.Args[0] {
			case "accept":
				handleAccept(ctx)
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
		ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("%s has won the wrestling duel against @%s and takes %d slices.", challenge.challengerName, ctx.Message.User.DisplayName, challenge.amount))
	} else {
		services.CurrencyServiceInstance.Give(challenge.challengerID, ctx.Message.User.ID, challenge.amount)
		ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("%s has won the wrestling duel against @%s and takes %d slices.", ctx.Message.User.DisplayName, challenge.challengerName, challenge.amount))
	}
}
