from discord.ext import commands
import discord
from bot.config import config
from bot.commands import loader
from bot.services import ollama_service
from bot.markov import markov
from bot.filter import filter


class DiscordChatMessageAdapter:
    """Make a Discord message behave like a Twitch ChatMessage for shared code."""

    def __init__(self, message: discord.Message):
        self.message = message
        self.text = message.content
        self.user = message.author

    async def reply(self, text: str):
        await self.message.reply(text, mention_author=True)

    async def send(self, text: str):
        await self.message.channel.send(text)


class DiscordChatCommand:
    """Wrap a Discord Context to match the interface of your Twitch ChatCommand."""

    def __init__(self, ctx, command_name, params=""):
        self.ctx = ctx
        self.name = command_name
        self.parameter = params
        self.author = ctx.author
        self.message = ctx.message

    async def send(self, content: str):
        await self.ctx.send(content)

    @property
    def user(self):
        """Mimic Twitch ChatCommand user object"""
        return self.author  # discord.Member or discord.User

    async def reply(self, content: str):
        """Reply to the user's message (like cmd.reply() on Twitch)."""
        await self.message.reply(content, mention_author=True)


class DiscordBot(commands.Bot):
    def __init__(self):
        intents: discord.Intents = discord.Intents.default()
        intents.message_content = True
        intents.members = True
        message_counter: int = 0
        super().__init__(command_prefix="!", intents=intents)

        async def hello(ctx):
            await ctx.send(f"Hello {ctx.author.mention}!")

        self.add_command(commands.Command(hello, name="hello", help="Says hello!"))

    async def setup_hook(self):
        print("[Discord] Setting up bot and loading commands...")
        commands_list: list[loader.BaseCommand] = await loader.load_commands()
        print(f"[Discord] Loaded {len(commands_list)} commands")

        for base_cmd in commands_list:
            # Capture properties
            name = base_cmd.name
            description = base_cmd.description
            usage = base_cmd.usage

            # Define a coroutine for the command
            async def command_coroutine(ctx, *args, cmd=base_cmd):
                # Join args into a parameter string
                params = " ".join(args)
                discord_cmd = DiscordChatCommand(ctx, cmd.name, params)
                try:
                    await cmd.execute(discord_cmd)
                except Exception as e:
                    await discord_cmd.send(f"Error executing `{cmd.name}`: {e}")
                    raise

            # Register the command dynamically
            self.add_command(
                commands.Command(
                    command_coroutine, name=name, help=description, usage=usage
                )
            )
            print(f"[Discord] Registered command: {name}")

        print("[Discord] All commands registered and bot setup complete.")

    async def on_ready(self):
        print(f"[Discord] Logged in as {self.user.name} (ID: {self.user.id})")

    async def on_message(self, message: discord.Message):
        # Ignore bot's own messages
        if message.author == self.user:
            return

        _config: config.Settings = config.get_settings()
        text = message.content
        if not text:
            return

        # 1️⃣ Handle commands
        if text.startswith("!"):
            await self.process_commands(message)
            return  # skip training/LLM for commands

        no_space_text = "".join(text.lower().split())

        # 2️⃣ Keyword responses (do not train)
        if "subathon" in no_space_text:
            if "jonathon" in no_space_text:
                await message.reply("fricc u")
            else:
                await message.reply("Jonathon* xddNerd")
            return  # skip training

        if "fricc" in no_space_text:
            await message.reply("fricc u too")
            return  # skip training

        # 3️⃣ Ignore specific users
        if message.author.name.lower() in _config.moderation.ignored_users:
            print(f"Ignoring {message.author.name}")
            return  # skip training

        # 4️⃣ Filter out links
        if filter.URL_REGEX.search(text):
            await message.reply("ben")
            return  # skip training

        # 5️⃣ Filter out bad words
        if filter.contains_badword(text, badwords=_config.moderation.bad_words):
            return  # skip training

        # ✅ If we reached this point, message passes all filters → log, Ollama, train
        print(f"{message.author.display_name}: {text}")

        # Pass to Ollama
        await ollama_service.ollama_service.on_message(
            msg=DiscordChatMessageAdapter(message)
        )

        # Markov training
        if _config.markov.train_on_chat:
            global message_counter
            await markov.build_ngrams(
                split_strategy=_config.markov.split_strategy,
                character_count=_config.markov.character_count,
                optional_text=text,
            )

            message_counter += 1
            if message_counter >= _config.markov.autosave_interval:
                message_counter = 0
                await markov.save_ngrams_to_binary(path=_config.markov.ngram_path)

    async def start_bot(self):
        token = config.get_settings().discord.token
        await self.start(token=token)
