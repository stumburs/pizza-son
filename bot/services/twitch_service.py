from twitchAPI.object.api import (
    ChannelInformation,
    TwitchUser,
    Moderator,
    ChannelVIP,
    Stream,
)
from bot.services import twitch_client
from twitchAPI.helper import first
from typing import AsyncGenerator
from bot.commands.base_command import PermissionLevel
from bot.config import config


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


async def get_stream_info(username: str) -> Stream | None:
    twitch = twitch_client.get_twitch()
    user = await get_user_by_name(username=username)

    stream: Stream = await first(twitch.get_streams(user_login=[user.login]))

    if not stream:
        print(f"No stream found with login: {username}")
        return None

    return stream


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


async def has_permissions(username: str, permissions: list[PermissionLevel]) -> bool:
    target_channel: str = config.get_settings().twitch.target_channel

    # All
    if PermissionLevel.ALL in permissions:
        return True

    # Streamer
    if (
        PermissionLevel.STREAMER in permissions
        and username.lower() == target_channel.lower()
    ):
        return True

    # Bot moderators
    if PermissionLevel.BOT_MODERATOR in permissions:
        bot_mods = config.get_settings().twitch.moderators
        if username.lower() in [m.lower() for m in bot_mods]:
            return True


# TODO: Figure out how to check permissions on other channels
async def _full_permission_check(
    username: str, permissions: list[PermissionLevel]
) -> bool:
    target_channel: str = config.get_settings().twitch.target_channel

    # All
    if PermissionLevel.ALL in permissions:
        return True

    # Streamer
    if (
        PermissionLevel.STREAMER in permissions
        and username.lower() == target_channel.lower()
    ):
        return True

    # Moderators
    if PermissionLevel.MODERATOR in permissions:
        moderators = await get_channel_moderators(target_channel)
        moderator_names = [mod.user_login.lower() for mod in moderators]
        if username.lower() in moderator_names:
            return True

    # VIPs
    if PermissionLevel.VIP in permissions:
        vips = await get_channel_vips(target_channel)
        vip_names = [vip.user_login.lower() for vip in vips]
        if username.lower() in vip_names:
            return True

    # Subscribers
    # TODO: implement a `get_channel_subscribers` helper
    if PermissionLevel.SUBSCRIBER in permissions:
        return False
        # subscribers = await get_channel_subscribers(target_channel)
        # sub_names = [sub.user_login.lower() for sub in subscribers]
        # if username.lower() in sub_names:
        #     return True

    # Bot moderators
    if PermissionLevel.BOT_MODERATOR in permissions:
        bot_mods = config.get_settings().twitch.moderators
        if username.lower() in [m.lower() for m in bot_mods]:
            return True
