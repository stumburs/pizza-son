from abc import ABC, abstractmethod
from twitchAPI.chat import ChatCommand


class BaseCommand(ABC):
    @property
    @abstractmethod
    def name(self) -> str:
        pass

    @property
    def description(self) -> str:
        return "No description provided."

    @abstractmethod
    async def execute(self, cmd: ChatCommand) -> None:
        pass
