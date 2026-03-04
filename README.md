# pizza-son

pizza's Twitch Chat Bot

## Requirements

- [Go](https://golang.org/dl/) 1.26+
- [Ollama](https://ollama.ai) (semi-optional (requires manually editing source code to disable), for `!llm`, `!nsfw`, etc. commands)

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/stumburs/pizza-son
cd pizza-son
```

### 2. Create your config file

Copy the example config and fill in your values:

```bash
cp config.example.toml config.toml
```

```toml
[twitch]
user = "your_bot_username"
oauth = "your_oauth_token"       # https://twitchtokengenerator.com/
client_id = "your_client_id"     # https://dev.twitch.tv/console
client_secret = "your_client_secret"

[bot]
prefix = "!"
channels = ["channel1", "channel2"]
ignored_users = ["streamelements", "nightbot"]

[ollama]
host = "http://localhost:11434"
model = "mistral:latest"
num_predict = 80
max_history = 80

[markov]
autosave_interval = 50
length_to_generate = 200
```

### 3. Set up Twitch credentials

1. Go to [dev.twitch.tv/console](https://dev.twitch.tv/console) and create a new application, ideally using your bot account.
2. Set the OAuth redirect URL to `http://localhost:3000`
3. Copy the **Client ID** and **Client Secret** into your config
4. Generate an OAuth token using [twitchtokengenerator.com](https://twitchtokengenerator.com/) and add it to your config **including `oauth:`**.

### 4. Set up the RNN model (semi-optional)

The `!rnn` command requires a pre-trained model. To train one from your own chat logs:

TODO: Add a script to automate this process

This produces `data/rnn/weights.json` and `data/rnn/vocab.json`. Training requires a `chat_logs.jsonl` file in the `tools/lm/` directory where each line is:

```json
{ "username": "user", "message": "hello", "timestamp": 1234567890 }
```

[!NOTE] You can disable this service by removing `services.NewRNNService()` from `main.go`.

### 5. Build and run

```bash
go build -o pizza-son ./cmd/pizza-son
./pizza-son
```

To cross-compile for Linux from Windows:

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o pizza-son ./cmd/pizza-son
```

---

## Commands

| Command                        | Description                                | Permission   |
| ------------------------------ | ------------------------------------------ | ------------ |
| `!ping`                        | Simple response test                       | Everyone     |
| `!hello [name]`                | Greets a user                              | Everyone     |
| `!8ball <question>`            | Ask the magic 8ball                        | Everyone     |
| `!quote`                       | Get a random quote                         | Everyone     |
| `!quote <number>`              | Get a specific quote                       | Everyone     |
| `!quote add <text>`            | Add a new quote                            | Moderator    |
| `!llm <prompt>`                | Chat with the LLM                          | Everyone     |
| `!mark`                        | Generate Markov chain text                 | Everyone     |
| `!rnn [seed]`                  | Generate text using the RNN model          | Everyone     |
| `!ada <message>`               | Chat with the learning bot                 | Everyone     |
| `!weather <location>`          | Get the weather                            | Everyone     |
| `!slices`                      | Check your pizza slice balance             | Everyone     |
| `!slices <user>`               | Check another user's balance               | Everyone     |
| `!slices give <user> <amount>` | Give slices to another user                | Everyone     |
| `!slices set <user> <amount>`  | Set a user's balance                       | Moderator    |
| `!mudwrestle <user> <amount>`  | Challenge someone to a mud wrestling match | Everyone     |
| `!mudwrestle accept`           | Accept a pending mud wrestling challenge   | Everyone     |
| `!lobotomize`                  | Clear the LLM chat context                 | Everyone     |
| `!perm`                        | Check your permission level                | Everyone     |
| `!reloadconfig`                | Reload config.toml                         | BotModerator |
| `!repeat <text>`               | Repeats what you say                       | Everyone     |

---

## Data

All persistent data is stored in the `data/` directory:

```
data/
├── markov/         # Per-channel Markov chain databases
├── ada/            # Per-channel Ada learning bot databases
├── quotes/         # Per-channel quote databases
├── logs/           # Per-channel message logs
├── currency/       # Global user balances
└── rnn/
    ├── weights.json
    └── vocab.json
```

> `data/rnn/weights.json` is not included in the repository due to its size. Generate it by following the RNN setup steps above.

---

## Adding Commands

Create a new file in `internal/commands/`:

```go
package commands

import "pizza-son/internal/bot"

func init() {
    Register(bot.Command{
        Name:        "mycommand",
        Description: "Does something cool.",
        Usage:       "!mycommand <arg>",
        Permission:  bot.All,
        Handler: func(ctx bot.CommandContext) {
            ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Hello!")
        },
    })
}
```

The command registers itself automatically via `init()` — no other files need to be changed.

## License is licensed under the MIT License. See [LICENSE](LICENSE) for details.
