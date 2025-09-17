from .base_command import BaseCommand
from twitchAPI.chat import ChatCommand
import services.twitch_service
import config.config as config


class GameCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "game"

    @property
    def description(self) -> str:
        return f"Replies with the current game being played on this stream. Usage !game"

    async def execute(self, cmd: ChatCommand) -> None:
        info = await services.twitch_service.get_channel_info(
            username=config.get_settings().twitch.target_channel
        )
        await cmd.reply(
            f"{info.broadcaster_name} is currently streaming {info.game_name}"
        )
