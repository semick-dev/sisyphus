from __future__ import annotations

from pathlib import Path

from agent.ado import logs
from agent.ado.client import ADOClient


BASE_TEMPLATE = """
Issue: {issue}

Instructions:
- We are attempting to resolve a failing build at
"""


FAILURE_TEMPLATE = """
Build Failure Detected
Build ID: {build_id}

Failed Log Excerpt (truncated):
{log_excerpt}
"""


def build_initial_instructions(*, issue: str, build_def: str, repo_path: Path) -> str:
    return BASE_TEMPLATE.format(
        issue=issue,
        build_def=build_def,
        repo_path=repo_path,
        instructions_path=repo_path / "instructions.md",
    )


def build_failure_instructions(
    *,
    issue: str,
    build_def: str,
    repo_path: Path,
    build_id: int,
    client: ADOClient,
    log_max_bytes: int,
) -> str:
    base = build_initial_instructions(issue=issue, build_def=build_def, repo_path=repo_path)
    log_excerpt = logs.fetch_failure_excerpt(client, build_id, max_bytes=log_max_bytes)
    return base + FAILURE_TEMPLATE.format(build_id=build_id, log_excerpt=log_excerpt)


def write_instructions(path: Path, content: str) -> None:
    path.write_text(content, encoding="utf-8")

