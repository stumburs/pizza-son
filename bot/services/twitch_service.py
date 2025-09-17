from twitchAPI.object.api import ChannelInformation, TwitchUser
import services.twitch_client as twitch_client
from twitchAPI.helper import first


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
