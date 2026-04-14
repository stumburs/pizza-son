package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
)

var pizzaTypes = []string{
	"Margherita",
	"Pepperoni",
	"BBQ Chicken",
	"Veggie",
	"Meat Lovers",
	"Buffalo Chicken",
	"Four Cheese",
	"Mushroom",
	"Supreme",
	"White Garlic",
	"Spinach & Feta",
	"Truffle",
	"Prosciutto",
	"Tuna & Onion",
	"Smoked Salmon",
	"Fig & Gorgonzola",
	"Caramelized Onion & Brie",
	"Pesto Chicken",
	"Chorizo & Jalapeño",
	"Mac & Cheese",
	"Cheeseburger",
	"Kebab",
	"Pulled Pork",
	"Teriyaki Chicken",
	"Kimchi",
	"Tandoori Chicken",
	"Nacho",
	"Breakfast",
	"Nutella & Banana",
	"Chocolate & Marshmallow",
	"Cookie Dough",
	"Strawberry & Cream",
	"Durian",
	"Anchovy",
	"Sardine & Olive",
	"Pineapple & Jalapeño",
	"Scorpion Pepper",
	"Ghost Pepper",
	"Charcoal Black",
	"Squid Ink",
	"Invisible",
	"Air",
	"Dirt",
	"Cardboard",
	"Mystery Meat",
	"Existential Dread",
	"Yesterday's Leftovers",
	"Suspiciously Cheap",
	"Definitely Not Dog Food",
}

func init() {
	Register(bot.Command{
		Name:        "eat",
		Description: "Eat a slice of pizza.",
		Usage:       "!eat",
		Category:    bot.CategoryFun,
		Examples: []bot.CommandExample{
			{Input: "!eat", Output: ""},
		},
		Handler: func(ctx bot.CommandContext) {
			// Exclude Discord
			if ctx.Message.Platform == models.PlatformDiscord {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "This is a Twitch exclusive command.")
				return
			}
			_, ok := services.CurrencyServiceInstance.Deduct(ctx.Message.User.ID, 1)
			if !ok {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "You don't have any pizza slices to eat... :(")
				return
			}
			pizza := pizzaTypes[rand.IntN(len(pizzaTypes))]
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("You ate a delicious slice of %s pizza.", pizza))
		},
	})
}
