import random
import pickle
import os

ngrams: dict[str, list[str]] = {}
SOURCE_TEXT: str = ""


async def load_ngrams_from_binary(path: str) -> None:
    global ngrams
    if not os.path.exists(path):
        with open(path, "wb") as f:
            pickle.dump({}, f)
        print(f"Ngrams file {path} did not exist, created empty file.")
        return

    with open(path, "rb") as f:
        ngrams = pickle.load(f)

    print(f"Ngrams successfully loaded from {path}!")


async def save_ngrams_to_binary(path: str) -> None:
    with open(path, "wb") as f:
        pickle.dump(ngrams, f)

    print(f"Ngrams successfully saved to {path}!")


async def get_random_key() -> str:
    if not ngrams:
        return "Ngrams is empty"

    return random.choice(list(ngrams.keys()))


async def generate_text(length: int) -> str:
    if not ngrams:
        return "No ngrams available - call build_ngrams() first!"

    current_word = random.choice(tuple(ngrams.keys()))
    result = [current_word]

    for _ in range(length):
        values = ngrams.get(current_word, [])
        if not values:
            break

        # Pick random next word
        current_word = random.choice(values)
        result.append(current_word)

    return "".join(result)


async def build_ngrams(
    split_strategy: str, character_count: int, optional_text: str | None = None
) -> str:
    text_to_use = optional_text if optional_text is not None else SOURCE_TEXT

    if split_strategy == "word":
        words = split_words_with_spaces(text_to_use)
    elif split_strategy == "character":
        words = split_by_n_characters(text_to_use, character_count)
        print(words)
    else:
        raise ValueError("Unknown split strategy to build Ngrams.")

    if len(words) == 1:
        word = words[0]
        if word not in ngrams:
            ngrams[word] = []
        print(f"Single {split_strategy} added to ngrams.")
        return

    for i in range(len(words) - 1):
        current_word = words[i]
        next_word = words[i + 1]

        if current_word not in ngrams:
            ngrams[current_word] = []

        ngrams[current_word].append(next_word)

    print(f"Ngrams built by {character_count} characters.")


# Helper function to split text by words (while keeping spaces)
def split_words_with_spaces(text: str) -> list[str]:
    """
    Splits text into words while keeping trailing spaces with the words.
    Example: "hello world" -> ["hello ", "world"]
    """
    chunks = []
    current_chunk = ""

    for char in text:
        current_chunk += char
        if (
            char.isspace() and current_chunk.strip()
        ):  # If we hit a space and have non-space chars
            chunks.append(current_chunk)
            current_chunk = ""

    # Add the last chunk if it's not empty
    if current_chunk:
        chunks.append(current_chunk)

    return chunks


# Helper function to split text by N characters
def split_by_n_characters(text: str, n: int) -> list[str]:
    result: list[str] = []

    for i in range(0, len(text), n):
        remaining: int = min(n, len(text) - i)
        result.append(text[i : i + remaining])

    return result
