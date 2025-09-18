from twitchAPI.chat import Chat
from twitchAPI.type import ChatEvent
from twitchAPI.oauth import UserAuthenticator
from twitchAPI.twitch import Twitch
from twitchAPI.type import AuthScope

import asyncio
from typing import List
from bot.token_store import token_store
from bot.commands import loader
from bot.on_message import on_message
from bot.on_ready import on_ready
from bot.markov import markov
from bot.services import ollama_service, twitch_service
from bot.config import config


async def run_bot() -> None:

    print("Loading config...")
    await config.reload_settings()
    settings: config.Settings = config.get_settings()

    if settings.ollama.host is not None:
        print("Initializing Ollama service...")
        ollama_service.ollama_service.init_client(settings=settings)

    print("Authenticating with Twitch...")
    bot = await Twitch(
        app_id=settings.twitch.client_id, app_secret=settings.twitch.client_secret
    )

    scopes: List[AuthScope] = [
        AuthScope.CHAT_READ,
        AuthScope.CHAT_EDIT,
        AuthScope.MODERATION_READ,
        AuthScope.CHANNEL_READ_VIPS,
    ]

    creds = token_store.load_tokens()
    if creds:
        token, refresh_token = creds
    else:
        auth = UserAuthenticator(bot, scopes=scopes)
        token, refresh_token = await auth.authenticate()
        token_store.save_tokens(token=token, refresh_token=refresh_token)

    await bot.set_user_authentication(
        token=token, scope=scopes, refresh_token=refresh_token
    )

    twitch_service.twitch_client.set_twitch(client=bot)

    chat = await Chat(bot)

    # Register events
    chat.register_event(ChatEvent.READY, on_ready)
    chat.register_event(ChatEvent.MESSAGE, on_message)

    # Register all commands
    print("Loading commands...")
    for command in loader.load_commands():
        chat.register_command(name=command.name, handler=command.execute)
        print(f"Registered command: {command.name}")

    chat.start()

    try:
        input("Press ENTER to stop! \n")
    finally:
        await markov.save_ngrams_to_binary(path=settings.markov.ngram_path)
        chat.stop()
        await bot.close()


if __name__ == "__main__":
    try:
        loop = asyncio.get_running_loop()
    except RuntimeError:  # no loop running
        loop = None

    if loop and loop.is_running():
        # We're in an environment with an already running loop
        asyncio.create_task(run_bot())
    else:
        # Normal script execution
        asyncio.run(run_bot())
