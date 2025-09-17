from twitchAPI.twitch import Twitch

_twitch: Twitch | None = None


def set_twitch(client: Twitch) -> None:
    global _twitch
    _twitch = client


def get_twitch() -> Twitch:
    if _twitch is None:
        raise RuntimeError("Twitch client not initialized yet.")
    return _twitch
