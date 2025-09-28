from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
from bot.services import ollama_service
from bot.services.twitch_service import has_permissions


class SetBestieVoiceCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "setvoice"

    @property
    def description(self) -> str:
        return "Sets the 'Bestie' TTS voice."

    @property
    def usage(self) -> str:
        return f"!{self.name} <voice>"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.STREAMER, PermissionLevel.BOT_MODERATOR]

    async def execute(self, cmd: ChatCommand) -> None:
        if not await has_permissions(cmd.user.name, self.permissions):
            return

        if not cmd.parameter:
            await cmd.reply(f"Usage: {self.usage}")
            return

        bestie_voice = cmd.parameter

        ollama_service.ollama_service.bestie_voice = bestie_voice

        await cmd.reply(f"Bestie has been set to {bestie_voice}!")
