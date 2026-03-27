package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strconv"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "prediction",
		Description: "Manage predictions using pizza slices.",
		Usage:       "!prediction start <title> | <option1> | <option2> ... | !prediction end <option> | !prediction cancel | !prediction info",
		Category:    bot.CategoryGames,
		Permission:  bot.Moderator,
		Examples: []bot.CommandExample{
			{Input: "!prediction start Will he win? | Yes | No", Output: "Prediction started: \"Will he win?\" - 1. Yes | 2. No | Bet with !bet <1/2> <amount>"},
			{Input: "!prediction end 1", Output: "Prediction ended! Winning option: Yes. Paying out winners..."},
			{Input: "!prediction cancel", Output: "Prediction cancelled. All bets refunded."},
			{Input: "!prediction info", Output: "\"Will he win?\" - 1. Yes (3 bets, 150 slices) | 2. No (2 bets, 80 slices)"},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !prediction start <title> | <opt1> | <opt2> | !prediction end <option> | !prediction cancel | !prediction info")
				return
			}
			switch strings.ToLower(ctx.Args[0]) {
			case "start":
				handlePredictionStart(ctx)
			case "end":
				handlePredictionEnd(ctx)
			case "cancel":
				handlePredictionCancel(ctx)
			case "info":
				handlePredictionInfo(ctx)
			default:
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !prediction start | end | cancel | info")
			}
		},
	})
	// Bet
	Register(bot.Command{
		Name:        "bet",
		Description: "Bet on currently ongoing predictions using pizza slices.",
		Usage:       "!bet <option> <amount>",
		Category:    bot.CategoryGames,
		Examples: []bot.CommandExample{
			{Input: "!bet 1 100", Output: "Bet 100 slices on option 1!"},
		},
		Handler: func(ctx bot.CommandContext) {
			p, ok := services.PredictionServiceInstance.GetActive(ctx.Message.Channel)
			if !ok {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No active predictions, as mods to start one.")
				return
			}

			if len(ctx.Args) < 2 {
				parts := make([]string, len(p.Outcomes))
				for i, o := range p.Outcomes {
					parts[i] = fmt.Sprintf("%d. %s", i+1, o.Title)
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID,
					fmt.Sprintf("\"%s\" - %s | Usage !bet <option> <amount>", p.Title, strings.Join(parts, " | ")))
				return
			}

			optionNum, err := strconv.Atoi(ctx.Args[0])
			if err != nil || optionNum < 1 || optionNum > len(p.Outcomes) {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID,
					fmt.Sprintf("Invalid option. Choose between 1 and %d.", len(p.Outcomes)))
				return
			}

			amount, err := strconv.Atoi(ctx.Args[1])
			if err != nil || amount <= 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Amount must be a positive number.")
				return
			}

			_, ok = services.CurrencyServiceInstance.Deduct(ctx.Message.User.ID, amount)
			if !ok {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID,
					fmt.Sprintf("You don't have enough pizza slices. Balance: %d", services.CurrencyServiceInstance.Balance(ctx.Message.User.ID)))
				return
			}

			outcome := p.Outcomes[optionNum-1]
			errMsg, success := services.PredictionServiceInstance.PlaceBet(ctx.Message.Channel, ctx.Message.User.ID, outcome.ID, amount)
			if !success {
				// Refund bet if failed
				services.CurrencyServiceInstance.Add(ctx.Message.User.ID, amount)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, errMsg)
				return
			}

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID,
				fmt.Sprintf("Bet %d slices on \"%s\"", amount, outcome.Title))
		},
	})
}

func handlePredictionStart(ctx bot.CommandContext) {
	if _, ok := services.PredictionServiceInstance.GetActive(ctx.Message.Channel); ok {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "A prediction is already active. End or cancel it first.")
		return
	}

	// Join all args split by |
	full := strings.Join(ctx.Args[1:], " ")
	parts := strings.Split(full, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	if len(parts) < 3 {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !prediction start <title> | <option1> | <option2> ...")
		return
	}

	title := parts[0]
	outcomeNames := parts[1:]

	if len(outcomeNames) > 10 {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Maximum 10 outcomes are allowed.")
		return
	}

	outcomes := make([]services.PredictionOutcome, len(outcomeNames))
	for i, name := range outcomeNames {
		outcomes[i] = services.PredictionOutcome{
			ID:    fmt.Sprintf("%d", i+1),
			Title: name,
		}
	}

	services.PredictionServiceInstance.Start(ctx.Message.Channel, ctx.Message.Channel, title, outcomes)

	optParts := make([]string, len(outcomes))
	for i, o := range outcomes {
		optParts[i] = fmt.Sprintf("%d. %s", i+1, o.Title)
	}
	ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("Prediction started: \"%s\" - %s | Bet with !bet <1-%d> <amount>",
		title, strings.Join(optParts, " | "), len(outcomes)))
}

func handlePredictionEnd(ctx bot.CommandContext) {
	p, ok := services.PredictionServiceInstance.GetActive(ctx.Message.Channel)
	if !ok {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No active prediction.")
		return
	}

	if len(ctx.Args) < 2 {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !prediction end <option number>")
		return
	}

	optionNum, err := strconv.Atoi(ctx.Args[1])
	if err != nil || optionNum < 1 || optionNum > len(p.Outcomes) {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Invalid option. Choose between 1 and %d.", len(p.Outcomes)))
		return
	}

	winningOutcome := p.Outcomes[optionNum-1]
	payouts := services.PredictionServiceInstance.End(ctx.Message.Channel, winningOutcome.ID)

	if len(payouts) == 0 {
		ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("Prediction ended! Winning option \"%s\". No bets were placed.", winningOutcome.Title))
		return
	}

	totalPot := 0
	for _, bet := range p.Bets {
		totalPot += bet.Amount

		ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("Prediction ended! Winning option \"%s\". %d slices got paid out to %d winner(s).",
			winningOutcome.Title, totalPot, len(payouts)))
	}
}

func handlePredictionCancel(ctx bot.CommandContext) {
	if _, ok := services.PredictionServiceInstance.GetActive(ctx.Message.Channel); !ok {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No active prediction.")
		return
	}
	services.PredictionServiceInstance.Cancel(ctx.Message.Channel)
	ctx.Client.Say(ctx.Message.Channel, "Prediction cancelled. All bets refunded.")
}

func handlePredictionInfo(ctx bot.CommandContext) {
	p, ok := services.PredictionServiceInstance.GetActive(ctx.Message.Channel)
	if !ok {
		ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No active prediction.")
		return
	}

	betCounts := make(map[string]int)
	betSlices := make(map[string]int)
	for _, bet := range p.Bets {
		betCounts[bet.OutcomeID]++
		betSlices[bet.OutcomeID] += bet.Amount
	}

	parts := make([]string, len(p.Outcomes))
	for i, o := range p.Outcomes {
		parts[i] = fmt.Sprintf("%d. %s (%d bets, %d slices)", i+1, o.Title, betCounts[o.ID], betSlices[o.ID])
	}
	ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("\"%s\" - %s", p.Title, strings.Join(parts, " | ")))
}
