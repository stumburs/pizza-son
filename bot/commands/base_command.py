from abc import ABC, abstractmethod
from twitchAPI.chat import ChatCommand


class BaseCommand(ABC):
    @property
    @abstractmethod
    def name(self) -> str:
        pass

    @abstractmethod
    async def execute(self, cmd: ChatCommand) -> None:
        pass
