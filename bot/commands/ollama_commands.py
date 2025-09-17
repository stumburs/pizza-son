from .base_command import BaseCommand
from twitchAPI.chat import ChatCommand
import services.ollama_service


def create_ollama_command(prompt_name: str):
    class OllamaPromptCommand(BaseCommand):
        @property
        def name(self) -> str:
            return prompt_name.lower()

        async def execute(self, cmd: ChatCommand) -> None:
            response = await services.ollama_service.ollama_service.get_llm_response(
                cmd
            )
            await cmd.reply(response)

    return OllamaPromptCommand()
