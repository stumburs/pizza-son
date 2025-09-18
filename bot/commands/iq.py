from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
import random


class PingCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "iq"

    @property
    def description(self) -> str:
        return "Shows the IQ of target user."

    @property
    def usage(self) -> str:
        return f"!hug [user]"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.ALL]

    async def execute(self, cmd: ChatCommand) -> None:
        iq_amount = random.randint(0, 300)

        if not cmd.parameter:
            await cmd.reply(f"{cmd.user.display_name} has {iq_amount} IQ xddNerd")
            return

        target = cmd.parameter.removeprefix("@").strip()
        await cmd.reply(f"{target} has {iq_amount} IQ xddNerd")
