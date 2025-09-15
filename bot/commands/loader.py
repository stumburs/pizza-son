import os
import importlib
from pathlib import Path
from typing import List
from .base_command import BaseCommand


def load_commands() -> List[BaseCommand]:
    commands = []
    commands_dir = Path(__file__).parent

    for file in commands_dir.glob("*.py"):
        if file.stem in ["__init__", "base_command", "loader"]:
            continue

        module_path = f"commands.{file.stem}"
        module = importlib.import_module(module_path)

        for attr_name in dir(module):
            attr = getattr(module, attr_name)
            if (
                isinstance(attr, type)
                and issubclass(attr, BaseCommand)
                and attr != BaseCommand
            ):
                commands.append(attr())

    return commands
