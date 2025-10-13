from twitchAPI.chat import ChatMessage
from .config import config
from .markov import markov
from bot.filter import filter
from bot.services import ollama_service
from bot.logging import logging

message_counter: int = 0


async def on_message(msg: ChatMessage) -> None:
    global message_counter

    _config: config.Settings = config.get_settings()

    # Log all messages
    if _config.logging.log_all_messages:
        logging.log_message(msg=msg)

    # filter out commands
    if msg.text.startswith("!"):
        return

    no_space_text = "".join(msg.text.lower().split())

    if "subathon" in no_space_text:
        if "jonathon" in no_space_text:
            await msg.reply("fricc u")
            return
        await msg.reply("Jonathon* xddNerd")

    if "fricc" in no_space_text:
        await msg.reply("fricc u too")

    # ignore specific users
    if msg.user.name.lower() in _config.moderation.ignored_users:
        print(f"Ignoring {msg.user.name}")
        return

    # filter out messages with links
    if filter.URL_REGEX.search(msg.text):
        await msg.reply("ben")
        return

    # filter out bad words
    if filter.contains_badword(msg.text, badwords=_config.moderation.bad_words):
        return

    print(f"{msg.user.display_name}: {msg.text}")

    await ollama_service.ollama_service.on_message(msg=msg)

    # train bot
    if _config.markov.train_on_chat:
        await markov.build_ngrams(
            split_strategy=_config.markov.split_strategy,
            character_count=_config.markov.character_count,
            optional_text=msg.text,
        )

        message_counter += 1

        if message_counter >= _config.markov.autosave_interval:
            message_counter = 0
            await markov.save_ngrams_to_binary(path=_config.markov.ngram_path)
