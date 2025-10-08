import tomllib
from typing import List
from pydantic import BaseModel
import aiofiles
from threading import Lock
from pathlib import Path

_lock = Lock()


class TwitchConfig(BaseModel):
    client_id: str
    client_secret: str
    target_channel: str
    moderators: List[str] = []


class DiscordConfig(BaseModel):
    token: str
    enabled: bool


class FeaturesConfig(BaseModel):
    tts: bool = False


class MarkovConfig(BaseModel):
    train_on_chat: bool = True
    ngram_path: str = "data.pkl"
    split_strategy: str = "character"
    character_count: int = 4
    autosave_interval: int = 50
    length_to_generate: int = 200
    max_retries: int = 3
    cooldown: int = 0


class ModerationConfig(BaseModel):
    ignored_users: List[str] = []
    bad_words: List[str] = []


class OllamaConfig(BaseModel):
    host: str = None
    model: str = None
    num_predict: int = 80
    max_history: int = 80


class AdaConfig(BaseModel):
    enabled: bool = False
    top_n: int = 3


class Settings(BaseModel):
    twitch: TwitchConfig
    discord: DiscordConfig
    features: FeaturesConfig
    markov: MarkovConfig
    moderation: ModerationConfig
    ollama: OllamaConfig


_settings: Settings | None = None

CONFIG_PATH: Path = Path("config.toml")


async def _load(config_path: str | Path = CONFIG_PATH) -> Settings:
    try:
        async with aiofiles.open(config_path, "rb") as f:
            content = await f.read()
            config_dict = tomllib.loads(content.decode("utf-8"))
        return Settings(**config_dict)

    except FileNotFoundError:
        raise FileNotFoundError(f"Configuration file not found: {config_path}")
    except tomllib.TOMLDecodeError as e:
        raise tomllib.TOMLDecodeError(f"Invalid TOML configuration: {str(e)}")


def get_settings() -> Settings:
    if _settings is None:
        raise RuntimeError("Settings not loaded yet. Call reload_settings() first.")
    return _settings


async def reload_settings(config_path: str | Path = CONFIG_PATH) -> Settings:
    global _settings
    _settings = await _load(config_path=CONFIG_PATH)
    return _settings
