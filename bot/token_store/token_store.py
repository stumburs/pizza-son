import json
from pathlib import Path
from typing import Optional, Tuple


TOKEN_FILE: Path = Path("tokens.json")


def load_tokens() -> Optional[Tuple[str, str]]:
    if not TOKEN_FILE.exists():
        return None
    data = json.loads(TOKEN_FILE.read_text())
    return data["token"], data["refresh_token"]


def save_tokens(token: str, refresh_token: str) -> None:
    tmp_file = TOKEN_FILE.with_suffix(".json.tmp")
    tmp_file.write_text(json.dumps({"token": token, "refresh_token": refresh_token}))
    tmp_file.replace(TOKEN_FILE)
