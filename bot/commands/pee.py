from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
import random


class PeeCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "pee"

    @property
    def description(self) -> str:
        return "Pees on the target person."

    @property
    def usage(self) -> str:
        return f"!{self.name} <user>"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.ALL]

    async def execute(self, cmd: ChatCommand) -> None:
        if not cmd.parameter:
            await cmd.reply(f"Usage: {self.usage}")
            return

        target = cmd.parameter.removeprefix("@").strip()
        flavor = random.randint(0, 100)

        await cmd.reply(
            f"{cmd.user.display_name} PEE s on borpaLick  {target} borpaLickL with {flavor}% flavor LICKA"
        )
