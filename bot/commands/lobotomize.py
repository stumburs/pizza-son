from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
from bot.services import ollama_service
from bot.services.twitch_service import has_permissions


class LobotomizeCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "lobotomize"

    @property
    def description(self) -> str:
        return f"Clears the AI's memory."

    @property
    def usage(self) -> str:
        return f"!{self.name}"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [
            PermissionLevel.STREAMER,
            PermissionLevel.MODERATOR,
            PermissionLevel.BOT_MODERATOR,
        ]

    async def execute(self, cmd: ChatCommand) -> None:
        if await has_permissions(cmd.user.name, self.permissions):
            await ollama_service.ollama_service.lobotomize()
            await cmd.reply("meow")
