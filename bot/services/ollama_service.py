from ollama import AsyncClient, ResponseError
from twitchAPI.chat import ChatMessage, ChatCommand
from collections import deque
from typing import Dict, Optional
import os
import config.config as config


class OllamaService:
    def __init__(self) -> None:
        self.client: Optional[AsyncClient] = None
        self.message_history: deque[Dict[str, str]] = deque(maxlen=80)
        self.system_prompts: dict[str, str] = {}
        self.model: str = ""
        self.num_predict: int = 80

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

    def get_system_prompt(self, prompt: str) -> dict[str, str]:
        return {
            "role": "system",
            "content": self.system_prompts.get(
                prompt, f"Respond with 'Failed to load prompt: {prompt}' only."
            ),
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

        messages = [self.get_system_prompt(cmd.name.lower())] + list(
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
