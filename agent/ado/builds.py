from __future__ import annotations

from urllib.parse import parse_qs, urlparse

from agent.ado.client import ADOClient


DEFAULT_API_VERSION = "7.1-preview.7"


def parse_build_url(build_url: str) -> tuple[str, str, str, str, str | None]:
    parsed = urlparse(build_url)
    if not parsed.scheme or not parsed.netloc:
        raise ValueError("Build URL must be a full URL with scheme and host.")

    host = parsed.netloc
    if host.endswith(".visualstudio.com"):
        org = host.split(".")[0]
        base_url = "https://dev.azure.com"
    else:
        base_url = f"{parsed.scheme}://{host}"
        org = parsed.path.strip("/").split("/")[0] if parsed.path.strip("/") else ""

    path_parts = [p for p in parsed.path.split("/") if p]
    if host.endswith(".visualstudio.com"):
        if len(path_parts) < 2:
            raise ValueError("Build URL path must include project name.")
        project = path_parts[0]
    else:
        if len(path_parts) < 3:
            raise ValueError("Build URL path must include org and project.")
        project = path_parts[1]

    qs = parse_qs(parsed.query)
    def_ids = qs.get("definitionId", [])
    build_ids = qs.get("buildId", [])

    build_def = def_ids[0] if def_ids else ""
    build_id = build_ids[0] if build_ids else None

    if not build_def and not build_id:
        raise ValueError("Build URL must include definitionId or buildId query param.")

    if not org or not project:
        raise ValueError("Could not parse org/project from build URL.")

    return org, project, build_def, base_url, build_id


def queue_build(client: ADOClient, definition: str, api_version: str = DEFAULT_API_VERSION) -> int:
    data = client.request(
        "POST",
        "/_apis/build/builds",
        params={"api-version": api_version},
        json={"definition": {"id": definition}},
    )
    return int(data["id"])


def get_build(client: ADOClient, build_id: int, api_version: str = DEFAULT_API_VERSION) -> dict:
    return client.request(
        "GET",
        f"/_apis/build/builds/{build_id}",
        params={"api-version": api_version},
    )


def get_build_status(client: ADOClient, build_id: int) -> str:
    data = get_build(client, build_id)
    return data.get("status", "unknown")


def get_build_result(client: ADOClient, build_id: int) -> str:
    data = get_build(client, build_id)
    return data.get("result", "unknown")


def get_build_definition_id(client: ADOClient, build_id: int) -> str:
    data = get_build(client, build_id)
    definition = data.get("definition", {})
    definition_id = definition.get("id")
    if definition_id is None:
        raise RuntimeError(f"Build {build_id} does not include a definition id.")
    return str(definition_id)
