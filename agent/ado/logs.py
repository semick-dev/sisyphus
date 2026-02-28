from __future__ import annotations

from agent.ado.client import ADOClient


DEFAULT_API_VERSION = "7.1-preview.2"


def list_logs(client: ADOClient, build_id: int, api_version: str = DEFAULT_API_VERSION) -> list[dict]:
    data = client.request(
        "GET",
        f"/_apis/build/builds/{build_id}/logs",
        params={"api-version": api_version},
    )
    return data.get("value", [])


def get_log(client: ADOClient, build_id: int, log_id: int, api_version: str = DEFAULT_API_VERSION) -> str:
    return client.request(
        "GET",
        f"/_apis/build/builds/{build_id}/logs/{log_id}",
        params={"api-version": api_version},
    )


def _truncate(text: str, max_bytes: int) -> str:
    encoded = text.encode("utf-8")
    if len(encoded) <= max_bytes:
        return text
    return encoded[:max_bytes].decode("utf-8", errors="ignore")


def fetch_failure_excerpt(client: ADOClient, build_id: int, max_bytes: int) -> str:
    logs = list_logs(client, build_id)
    if not logs:
        return "<no logs available>"
    last_log_id = logs[-1]["id"]
    content = get_log(client, build_id, last_log_id)
    return _truncate(content, max_bytes)
