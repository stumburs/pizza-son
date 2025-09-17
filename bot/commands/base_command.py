from abc import ABC, abstractmethod
from twitchAPI.chat import ChatCommand

from enum import Enum


class PermissionLevel(str, Enum):
    ALL = "All"
    VIP = "VIP"
    SUBSCRIBER = "Subscriber"
    MODERATOR = "Moderator"
    STREAMER = "Streamer"
    BOT_MODERATOR = "Bot Moderator"


class BaseCommand(ABC):
    @property
    @abstractmethod
    def name(self) -> str:
        pass

    @property
    def description(self) -> str:
        return "No description provided."

    @property
    def usage(self) -> str:
        return "No usage provided."

    @property
    def permissions(self) -> list[PermissionLevel]:
        return [PermissionLevel.ALL]

    @abstractmethod
    async def execute(self, cmd: ChatCommand) -> None:
        pass
