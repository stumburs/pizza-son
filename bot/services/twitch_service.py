from twitchAPI.object.api import ChannelInformation, TwitchUser, Moderator, ChannelVIP
from bot.services import twitch_client
from twitchAPI.helper import first
from typing import AsyncGenerator


async def get_user_by_name(username: str) -> TwitchUser:
    twitch = twitch_client.get_twitch()
    users = await first(twitch.get_users(logins=[username]))
    if not users:
        raise ValueError(f"No user found with login: {username}")
    return users


async def get_channel_info(username: str) -> ChannelInformation:
    twitch = twitch_client.get_twitch()
    user = await get_user_by_name(username=username)
    channels = await twitch.get_channel_information(broadcaster_id=user.id)
    if not channels:
        raise ValueError(f"No channel info found for user: {username}")
    return channels[0]


async def get_channel_moderators(username: str) -> list[Moderator]:
    twitch = twitch_client.get_twitch()
    broadcaster_id: str = (await get_user_by_name(username=username)).id
    moderators_async_gen: AsyncGenerator[Moderator, None] = twitch.get_moderators(
        broadcaster_id=broadcaster_id
    )
    moderators: list[Moderator] = [item async for item in moderators_async_gen]
    return moderators


async def get_channel_vips(username: str) -> list[ChannelVIP]:
    twitch = twitch_client.get_twitch()
    broadcaster_id: str = (await get_user_by_name(username=username)).id
    vip_async_gen: AsyncGenerator[ChannelVIP, None] = twitch.get_vips(
        broadcaster_id=broadcaster_id
    )
    vips: list[Moderator] = [item async for item in vip_async_gen]
    return vips
