from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
import random


class PingCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "hug"

    @property
    def description(self) -> str:
        return "Hugs the target user."

    @property
    def usage(self) -> str:
        return f"!hug <user>"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.ALL]

    async def execute(self, cmd: ChatCommand) -> None:
        if not cmd.parameter:
            await cmd.reply(f"Usage: {self.usage()}")
            return

        target = cmd.parameter.removeprefix("@").strip()
        love = random.randint(0, 100)

        await cmd.reply(f"{cmd.user.display_name} hugs {target} with {love}% love :3")
