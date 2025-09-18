from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
from bot.services import twitch_service
from bot.config import config


class GameCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "game"

    @property
    def description(self) -> str:
        return f"Replies with the current game being played on this stream."

    @property
    def usage(self) -> str:
        return f"!{self.name()}"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.ALL]

    async def execute(self, cmd: ChatCommand) -> None:
        info = await twitch_service.get_channel_info(
            username=config.get_settings().twitch.target_channel
        )
        await cmd.reply(
            f"{info.broadcaster_name} is currently streaming {info.game_name}"
        )
