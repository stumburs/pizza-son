import json
from pathlib import Path
from twitchAPI.chat import ChatMessage
from bot.config import config

LOG_FILE = Path("logs/chat_logs.jsonl")
LOG_FILE.parent.mkdir(exist_ok=True)


async def log_message(msg: ChatMessage) -> None:
    try:
        target_channel = config.get_settings().twitch.target_channel

        log_data = {
            "username": msg.user.name,
            "message": msg.text,
            "timestamp": msg.sent_timestamp,
            "streamer": target_channel,
            "first_msg": msg.first,
            "subscriber": msg.user.subscriber,
            "vip": msg.user.vip,
            "mod": msg.user.mod,
        }

        with LOG_FILE.open("a", encoding="utf-8") as f:
            f.write(json.dumps(log_data, ensure_ascii=False) + "\n")

    except Exception as e:
        print(f"[Logging] Logging error: {e}")
