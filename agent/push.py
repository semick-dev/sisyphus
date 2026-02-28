import subprocess
import time
from pathlib import Path

from agent import payload
from agent.ado import builds
from agent.ado import client as ado_client


def _run_cmd(cmd: list[str], cwd: Path | None = None) -> None:
    result = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)
    if result.returncode != 0:
        raise RuntimeError(
            "Command failed: "
            + " ".join(cmd)
            + "\nstdout:\n"
            + result.stdout
            + "\nstderr:\n"
            + result.stderr
        )


def _git_status(repo_path: Path) -> str:
    result = subprocess.run(
        ["git", "status", "--porcelain"],
        cwd=repo_path,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(f"git status failed: {result.stderr.strip()}")
    return result.stdout


def _ensure_clean(repo_path: Path) -> None:
    status = _git_status(repo_path).strip()
    if status:
        raise RuntimeError("Working tree is dirty. Commit or stash before running.")


def _ensure_has_changes(repo_path: Path) -> None:
    status = _git_status(repo_path).strip()
    if not status:
        raise RuntimeError("No changes to commit after LLM run.")


def _invoke_llm(llm: str, instructions_path: Path) -> None:
    if llm == "codex":
        cmd = ["codex", "-p", str(instructions_path), "--autopilot"]
    elif llm == "claude":
        cmd = ["claude", "-p", str(instructions_path)]
    elif llm == "copilot":
        cmd = ["copilot", "-p", str(instructions_path)]
    else:
        raise ValueError(f"Unsupported llm: {llm}")

    _run_cmd(cmd)


def _git_commit_and_push(repo_path: Path, message: str) -> None:
    _run_cmd(["git", "add", "-A"], cwd=repo_path)
    _ensure_has_changes(repo_path)
    _run_cmd(["git", "commit", "-m", message], cwd=repo_path)
    _run_cmd(["git", "push"], cwd=repo_path)


def run(
    *,
    issue: str,
    build_def: str,
    start_build_id: int | None,
    repo_path: Path,
    llm: str,
    sleep_seconds: int,
    log_max_bytes: int,
    ado_org: str,
    ado_project: str,
    ado_base_url: str,
    pat: str,
) -> int:
    instructions_path = repo_path / "instructions.md"

    _ensure_clean(repo_path)

    client = ado_client.ADOClient(
        org=ado_org,
        project=ado_project,
        base_url=ado_base_url,
        pat=pat,
    )

    build_id = start_build_id
    effective_build_def = build_def
    if build_id is None:
        base_instructions = payload.build_initial_instructions(
            issue=issue,
            build_def=build_def,
            repo_path=repo_path,
        )
        payload.write_instructions(instructions_path, base_instructions)

        _invoke_llm(llm, instructions_path)

        _git_commit_and_push(repo_path, "Automated agent update")

        if not effective_build_def:
            raise RuntimeError("Missing build definition id; cannot queue a new build.")
        build_id = builds.queue_build(client, effective_build_def)
        time.sleep(sleep_seconds)
    else:
        if not effective_build_def:
            effective_build_def = builds.get_build_definition_id(client, build_id)

    while True:
        status = builds.get_build_status(client, build_id)
        if status == "completed":
            result = builds.get_build_result(client, build_id)
            if result == "succeeded":
                return 0

            failure_payload = payload.build_failure_instructions(
                issue=issue,
                build_def=effective_build_def,
                repo_path=repo_path,
                build_id=build_id,
                client=client,
                log_max_bytes=log_max_bytes,
            )
            payload.write_instructions(instructions_path, failure_payload)
            _invoke_llm(llm, instructions_path)
            _git_commit_and_push(repo_path, "Automated fix for build failure")
            if not effective_build_def:
                raise RuntimeError("Missing build definition id; cannot queue a new build.")
            build_id = builds.queue_build(client, effective_build_def)
            time.sleep(sleep_seconds)
            continue

        time.sleep(sleep_seconds)
