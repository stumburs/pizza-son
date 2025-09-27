from pathlib import Path
from bot.commands.loader import load_commands
from bot.commands.base_command import PermissionLevel
import asyncio


async def main():
    docs_path = Path("docs/commands.md")
    docs_path.parent.mkdir(exist_ok=True)

    commands = await load_commands()
    print("Loaded commands:", [cmd.name for cmd in commands])

    with docs_path.open("w", encoding="utf-8") as f:
        f.write("# pizza_son Commands\n\n")
        f.write("This page lists all* available bot commands.\n\n")

        for cmd in commands:
            name = f"!{cmd.name}"
            description = getattr(cmd, "description", "No description provided.")
            usage = getattr(cmd, "usage", f"!{cmd.name}")
            permissions = getattr(cmd, "permissions", [PermissionLevel.ALL])

            if isinstance(permissions, list):
                perm_str = ", ".join(
                    [
                        p.value if isinstance(p, PermissionLevel) else str(p)
                        for p in permissions
                    ]
                )
            else:
                perm_str = str(permissions)

            f.write(f"## {name}\n\n")
            f.write(f"**Description:**\n\n{description}\n\n")
            f.write(f"**Usage:**\n\n```\n{usage}\n```\n\n")
            f.write(f"**Permissions:**\n\n{perm_str}\n\n")
            f.write("---\n\n")

    print("Generated docs/commands.md")


if __name__ == "__main__":
    asyncio.run(main())
