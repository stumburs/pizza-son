from twitchAPI.chat import EventData
import config.config as config
import markov.markov as markov


async def on_ready(ready_event_data: EventData) -> None:
    settings: config.Settings = config.get_settings()

    print("Loading markov data...")
    await markov.load_ngrams_from_binary(settings.markov.ngram_path)

    await ready_event_data.chat.join_room(settings.twitch.target_channel)
    print(f"Bot successfully connected to {settings.twitch.target_channel}'s chat!")
