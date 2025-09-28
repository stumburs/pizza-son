from ollama import AsyncClient, ResponseError
from twitchAPI.chat import ChatMessage, ChatCommand
from collections import deque
from typing import Dict, Optional
import os
from bot.config import config
from bot.services import twitch_service
import re
from datetime import datetime
import pytz


class OllamaService:
    def __init__(self) -> None:
        self.client: Optional[AsyncClient] = None
        self.message_history: deque[Dict[str, str]] = deque(maxlen=80)
        self.system_prompts: dict[str, str] = {}
        self.model: str = ""
        self.num_predict: int = 80
        self.bestie_enabled: bool = False
        self.bestie_voice: str = "bestie"

    def init_client(self, settings: config.Settings) -> None:
        self.client = AsyncClient(host=settings.ollama.host)
        self.system_prompts = self._load_system_prompts()
        self.model = settings.ollama.model
        self.num_predict = settings.ollama.num_predict

    def _load_system_prompts(self) -> dict[str, str]:
        prompts: dict[str, str] = {}
        prompts_dir = os.path.join(os.path.dirname(__file__), "ollama_prompts")

        if not os.path.exists(prompts_dir):
            os.makedirs(prompts_dir)

        for filename in os.listdir(prompts_dir):
            if filename.endswith(".txt"):
                prompt_name = os.path.splitext(filename)[0]
                with open(
                    os.path.join(prompts_dir, filename), "r", encoding="utf-8"
                ) as f:
                    prompts[prompt_name] = f.read().strip()

        return prompts

    def list_prompts_without_client(settings) -> list[str]:
        temp_client = OllamaService()
        temp_client.init_client(settings=settings)
        return temp_client.get_available_prompts()

    async def fill_placeholders(self, text: str) -> str:
        target_channel = config.get_settings().twitch.target_channel
        channel_info = await twitch_service.get_channel_info(target_channel)

        stream_info = await twitch_service.get_stream_info(target_channel)

        berlin_tz = pytz.timezone("Europe/Berlin")
        time_cest = datetime.now(berlin_tz).strftime("%Y-%m-%d %H:%M:%S %Z")

        replacements = {
            "game_name": channel_info.game_name,
            "stream_title": channel_info.title,
            "channel_name": target_channel,
            "time_cest": time_cest,
            "channel_tags": ", ".join(channel_info.tags),
            "viewer_count": (
                stream_info.viewer_count if stream_info else "Failed to fetch stream."
            ),
            "thumbnail_url": (
                stream_info.thumbnail_url if stream_info else "Failed to fetch stream."
            ),
        }

        def replacer(match):
            key = match.group(1)
            return replacements.get(key, f"{{{{{key}}}}}")

        return re.sub(r"\{\{(\w+)\}\}", replacer, text)

    async def get_system_prompt(self, prompt: str) -> dict[str, str]:
        base_prompt = self.system_prompts.get(
            prompt, f"Respond with 'Failed to load prompt: {prompt}' only."
        )

        content = await self.fill_placeholders(base_prompt)

        return {
            "role": "system",
            "content": content,
        }

    async def on_message(self, msg: ChatMessage) -> None:
        self.message_history.append(
            {"role": "user", "content": f"{msg.user.display_name} chatted: {msg.text}"}
        )
        print(f"[OllamaService] Recorded message: {self.message_history[-1]}")

    async def lobotomize(self) -> None:
        self.message_history.clear()
        print("[OllamaService] Message history cleared.")

    async def get_llm_response(self, cmd: ChatCommand) -> str:
        if self.client is None:
            return "Ollama client not initialized."

        question = {
            "role": "user",
            "content": f"{cmd.user.display_name} asked assistant: {cmd.parameter}",
        }

        self.message_history.append(question)
        print(f"[OllamaService] Question: {question}")

        messages = [await self.get_system_prompt(cmd.name.lower())] + list(
            self.message_history
        )

        try:
            response = await self.client.chat(
                model=self.model,
                messages=messages,
                options={"num_predict": self.num_predict},
            )

            self.message_history.append(
                {"role": "assistant", "content": response.message.content}
            )

            print(f"[OllamaService] Response: {response.message.content}")

            return response.message.content[:400]
        except ResponseError as e:
            return f"Something went wrong while generating response: {e.error}"

    def get_available_prompts(self) -> list[str]:
        return list(self.system_prompts.keys())


# Global instance
ollama_service = OllamaService()
