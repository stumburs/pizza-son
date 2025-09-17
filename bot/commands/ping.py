from .base_command import BaseCommand
from twitchAPI.chat import ChatCommand


class PingCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "ping"

    @property
    def description(self) -> str:
        return "Replies with 'pong!' to test the bot's responsiveness. Usage: !ping"

    async def execute(self, cmd: ChatCommand) -> None:
        await cmd.reply("pong!")
