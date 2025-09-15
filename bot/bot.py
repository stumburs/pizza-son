from twitchAPI.chat import Chat, EventData, ChatMessage, ChatCommand
from twitchAPI.type import ChatEvent
from twitchAPI.oauth import UserAuthenticator
from twitchAPI.twitch import Twitch
from twitchAPI.type import AuthScope


import asyncio
from typing import List
import config.config as config
import token_store.token_store as token_store


async def on_ready(ready_event_data: EventData) -> None:
    settings: config.Settings = config.get_settings()
    await ready_event_data.chat.join_room(settings.twitch.target_channel)
    print(f"Bot successfully connected to {settings.twitch.target_channel}'s chat!")


async def on_message(msg: ChatMessage) -> None:
    print(f"{msg.user.display_name}: {msg.text}")


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

    chat.register_event(ChatEvent.READY, on_ready)
    chat.register_event(ChatEvent.MESSAGE, on_message)

    chat.start()

    try:
        input("Press ENTER to stop! \n")
    finally:
        chat.stop()
        await bot.close()


asyncio.run(run_bot())
