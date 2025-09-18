from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand


class GetDocsCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "commands"

    @property
    def description(self) -> str:
        return f"Sends a link to this page."

    @property
    def usage(self) -> str:
        return f"!{self.name()}"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.ALL]

    async def execute(self, cmd: ChatCommand) -> None:
        await cmd.reply(
            "Use and abuse me with these commands: https://stumburs.github.io/pizza-son/commands/"
        )
