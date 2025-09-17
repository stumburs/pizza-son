from pathlib import Path
from bot.commands.loader import load_commands

docs_path = Path("docs/commands.md")

with docs_path.open("w", encoding="utf-8") as f:
    f.write("# pizza_son commands\n\n")
    for cmd in load_commands():
        f.write(f"## !{cmd.name}\n")
        f.write(f"Description: {getattr(cmd, 'description', 'No description')}\n\n")

print("Generated docs/commands.md")
