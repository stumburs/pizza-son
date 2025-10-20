import importlib
from pathlib import Path
from typing import List
from .base_command import BaseCommand
from bot.config import config


async def load_commands() -> List[BaseCommand]:
    commands = []
    commands_dir = Path(__file__).parent

    for file in commands_dir.glob("*.py"):
        if file.stem in ["__init__", "base_command", "loader", "ollama_commands"]:
            continue

        module_path = f"bot.commands.{file.stem}"
        module = importlib.import_module(module_path)

        for attr_name in dir(module):
            attr = getattr(module, attr_name)
            if (
                isinstance(attr, type)
                and issubclass(attr, BaseCommand)
                and attr != BaseCommand
            ):
                commands.append(attr())

    from bot.services.ollama_service import ollama_service, OllamaService
    from bot.commands.ollama_commands import create_ollama_command

    await config.reload_settings()

    prompts: list[str] = []

    # if global service is ready, use it
    if ollama_service.client is not None:
        prompts = ollama_service.get_available_prompts()
    else:
        # fallback for docs generation
        prompts = OllamaService.list_prompts_without_client(config.get_settings())

    for prompt in prompts:
        if prompt in ["lark", "jark"]:
            continue
        commands.append(create_ollama_command(prompt_name=prompt))

    return commands
