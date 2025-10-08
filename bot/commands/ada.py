from .base_command import BaseCommand, PermissionLevel
from twitchAPI.chat import ChatCommand
from bot.ada import ada


class AdaCommand(BaseCommand):
    @property
    def name(self) -> str:
        return "ada"

    @property
    def description(self) -> str:
        return "Responds based off previous interactions. Functions similarly to the chatbot - [Cleverbot](https://www.cleverbot.com/)."

    @property
    def usage(self) -> str:
        return f"!{self.name} <text>"

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.ALL]

    async def execute(self, cmd: ChatCommand) -> None:
        if not cmd.parameter:
            await cmd.reply(f"Usage: {self.usage}")
            return

        response = await ada.get_ada_response(user_input=cmd.parameter)

        await cmd.reply(response)
