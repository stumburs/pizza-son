from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
from bot.services import ollama_service
from bot.services.twitch_service import has_permissions


class BestieToggleCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "togglebestie"

    @property
    def description(self) -> str:
        return "Toggles the 'Bestie' TTS for LLM responses."

    @property
    def usage(self) -> str:
        return f"!{self.name}"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.STREAMER, PermissionLevel.BOT_MODERATOR]

    async def execute(self, cmd: ChatCommand) -> None:
        if not await has_permissions(cmd.user.name, self.permissions):
            return

        ollama_service.ollama_service.bestie_enabled = (
            not ollama_service.ollama_service.bestie_enabled
        )

        await cmd.reply(
            f"Bestie has been {'enabled' if ollama_service.ollama_service.bestie_enabled else 'disabled'}!"
        )
