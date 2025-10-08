import json, os, random, asyncio, functools
import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity

DATA_FILE = "conversations.json"

if os.path.exists(DATA_FILE):
    with open("conversations.json", "r", encoding="utf-8") as f:
        database = json.load(f)
else:
    database = [
        {"input": "Hi!", "response": "hello! how are you?"},
        {"input": "Hello!", "response": "hi there!"},
        {"input": "How are you?", "response": "I'm fine, thanks for asking!"},
        {"input": "What's your name?", "response": "I'm Ada, your lovely chatbot."},
        {"input": "What do you like?", "response": "I like chatting with you!"},
    ]


async def save_database():
    loop = asyncio.get_event_loop()
    await loop.run_in_executor(None, functools.partial(_save_database_blocking))


def _save_database_blocking():
    with open(DATA_FILE, "w", encoding="utf-8") as f:
        json.dump(database, f, indent=2, ensure_ascii=False)


def build_vectorizer():
    inputs = [pair["input"] for pair in database]
    vectorizer = TfidfVectorizer().fit(inputs)
    input_vectors = vectorizer.transform(inputs)
    return vectorizer, input_vectors


vectorizer, input_vectors = build_vectorizer()


def find_best_response(user_message: str, top_n=3):
    user_vec = vectorizer.transform([user_message])
    sims = cosine_similarity(user_vec, input_vectors).flatten()

    sorted_indices = np.argsort(sims)[::-1]
    top_indices = sorted_indices[:top_n]
    top_scores = sims[top_indices]

    if np.max(top_scores) < 0.2:
        return None, np.max(top_scores)

    total = np.sum(top_scores)
    if total == 0:
        probs = np.ones_like(top_scores) / len(top_scores)
    else:
        probs = top_scores / total

    chosen_idx = np.random.choice(top_indices, p=probs)
    score = sims[chosen_idx]

    if score < 0.2:
        return None, score

    return database[chosen_idx]["response"], score


print("[Ada] Now active and learning!")

last_bot_message = None


async def get_ada_response(user_input: str) -> str:
    global last_bot_message, database, vectorizer, input_vectors

    user_input = user_input.strip()
    if not user_input:
        return None

    if last_bot_message:
        new_pair = {"input": last_bot_message, "response": user_input}
        database.append(new_pair)
        vectorizer, input_vectors = build_vectorizer()
        await save_database()

    reply, score = find_best_response(user_message=user_input)
    if reply:
        last_bot_message = reply
        print(f"[Ada] - User asked: {user_input}, response:{reply} ({score})")
        return reply
    else:
        fallback = random.choice(
            ["Interesting!", "Tell me more.", "Hmm, okay.", "Why do you say that?"]
        )
        return fallback
