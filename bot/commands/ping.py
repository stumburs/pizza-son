from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand


class PingCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "ping"

    @property
    def description(self) -> str:
        return "Replies with 'pong!' to test the bot's responsiveness."

    @property
    def usage(self) -> str:
        return f"!{self.name()}"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.ALL]

    async def execute(self, cmd: ChatCommand) -> None:
        await cmd.reply("pong!")
