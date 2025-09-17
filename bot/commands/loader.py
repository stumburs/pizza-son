import importlib
from pathlib import Path
from typing import List
from .base_command import BaseCommand


def load_commands() -> List[BaseCommand]:
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

    from bot.services.ollama_service import ollama_service
    from bot.commands.ollama_commands import create_ollama_command

    for prompt in ollama_service.get_available_prompts():
        commands.append(create_ollama_command(prompt_name=prompt))

    return commands
