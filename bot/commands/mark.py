from .base_command import BaseCommand
from twitchAPI.chat import ChatCommand
import markov.markov as markov
import config.config as config


class MarkCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "mark"

    async def execute(self, cmd: ChatCommand) -> None:
        length_to_generate: int = config.get_settings().markov.length_to_generate
        await cmd.reply(await markov.generate_text(length=length_to_generate))
