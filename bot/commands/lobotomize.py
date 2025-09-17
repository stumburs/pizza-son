from .base_command import BaseCommand
from twitchAPI.chat import ChatCommand
from services.ollama_service import ollama_service


class LobotomizeCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "lobotomize"

    async def execute(self, cmd: ChatCommand) -> None:
        await ollama_service.lobotomize()
        await cmd.reply("meow")
