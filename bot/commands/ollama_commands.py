from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
from bot.services import ollama_service


def create_ollama_command(prompt_name: str):
    class OllamaPromptCommand(BaseCommand):
        @property
        def name(self) -> str:
            return prompt_name.lower()

        @property
        def description(self) -> str:
            return f"Replies as '{prompt_name} using an AI."

        @property
        def usage(self) -> str:
            return f"Usage: !{prompt_name.lower()} <question>"

        @property
        def permissions(self) -> list[PermissionLevel]:
            return [PermissionLevel.ALL]

        async def execute(self, cmd: ChatCommand) -> None:
            response = await ollama_service.ollama_service.get_llm_response(cmd)
            await cmd.reply(response)

    return OllamaPromptCommand()
