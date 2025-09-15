from twitchAPI.chat import Chat, EventData, ChatMessage, ChatCommand
from twitchAPI.type import ChatEvent
from twitchAPI.oauth import UserAuthenticator
from twitchAPI.twitch import Twitch
from twitchAPI.type import AuthScope


import asyncio
from typing import List
import config.config as config
import token_store.token_store as token_store
import commands.loader
import on_message
import on_ready
import markov.markov as markov


async def run_bot() -> None:

    print("Loading config...")
    await config.reload_settings()

    settings: config.Settings = config.get_settings()

    bot = await Twitch(
        app_id=settings.twitch.client_id, app_secret=settings.twitch.client_secret
    )

    print("Authenticating with Twitch...")
    scopes: List[AuthScope] = [AuthScope.CHAT_READ, AuthScope.CHAT_EDIT]
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

    chat = await Chat(bot)

    # Register events
    chat.register_event(ChatEvent.READY, on_ready.on_ready)
    chat.register_event(ChatEvent.MESSAGE, on_message.on_message)

    # Register all commands
    print("Loading commands...")
    for command in commands.loader.load_commands():
        chat.register_command(name=command.name, handler=command.execute)
        print(f"Registered command: {command.name}")

    chat.start()

    try:
        input("Press ENTER to stop! \n")
    finally:
        await markov.save_ngrams_to_binary(path=settings.markov.ngram_path)
        chat.stop()
        await bot.close()


asyncio.run(run_bot())
