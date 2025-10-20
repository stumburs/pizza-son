from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
from bot.config import config
from bot.markov import markov
from bot.services import ollama_service


class LarkCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "lark"

    @property
    def description(self) -> str:
        return "Generates a random message trained on this chat (!mark) and reinterprets the output using an LLM."

    @property
    def usage(self) -> str:
        return f"!{self.name}"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.ALL]

    async def execute(self, cmd: ChatCommand) -> None:
        length_to_generate: int = config.get_settings().markov.length_to_generate
        mark_result: str = await markov.generate_text(length=length_to_generate)

        temp_cmd: ChatCommand = cmd
        temp_cmd.parameter = mark_result

        llm_response: str = (
            await ollama_service.ollama_service.get_llm_response_without_memory(
                cmd=temp_cmd
            )
        )

        await cmd.reply(llm_response)
