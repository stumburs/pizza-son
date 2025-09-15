from twitchAPI.chat import ChatMessage
import config.config as config
import markov.markov as markov
import filter.filter as filter

message_counter: int = 0


async def on_message(msg: ChatMessage) -> None:
    global message_counter

    _config: config.Settings = config.get_settings()

    # filter out commands
    if msg.text.startswith("!"):
        return

    if "subathon" in msg.text.lower():
        if "jonathon" in msg.text.lower():
            await msg.reply("fricc u")
            return
        await msg.reply("Jonathon* xddNerd")

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

    # train bot
    if _config.markov.train_on_chat:
        print(
            await markov.build_ngrams(
                split_strategy=_config.markov.split_strategy,
                character_count=_config.markov.character_count,
                optional_text=msg.text,
            )
        )

        message_counter += 1

        if message_counter >= _config.markov.autosave_interval:
            message_counter = 0
            await markov.save_ngrams(path=_config.markov.ngram_path)
