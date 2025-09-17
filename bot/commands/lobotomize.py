from .base_command import BaseCommand
from twitchAPI.chat import ChatCommand
from bot.services import ollama_service


class LobotomizeCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "lobotomize"

    @property
    def description(self) -> str:
        return f"Clears the AI's memory. Note: This command is only available to moderators listed in the bots configuration file. Usage !lobotomize"

    async def execute(self, cmd: ChatCommand) -> None:
        await ollama_service.lobotomize()
        await cmd.reply("meow")
