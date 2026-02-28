import argparse
import os
import subprocess
import sys
from pathlib import Path

from agent import push
from agent.ado import builds


def _parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="sisypnus-agent",
        description="Run the Sisypnus agent loop for issue-driven builds.",
    )
    parser.add_argument(
        "--issue",
        required=True,
        help="Issue identifier (Org/repo#xxx) or starting prompt.",
    )
    parser.add_argument(
        "--build",
        required=True,
        help=(
            "ADO build definition URL or build results URL. "
            "Definition example: https://org.visualstudio.com/Project/_build?definitionId=1. "
            "Results example: https://org.visualstudio.com/Project/_build/results?buildId=447&view=results."
        ),
    )
    parser.add_argument(
        "--pat",
        required=False,
        help="ADO PAT token. Optionally sourced from ADO_PAT environment variable.",
    )
    parser.add_argument(
        "--llm",
        default="codex",
        choices=["codex", "claude", "copilot"],
        help="LLM CLI to invoke for autopilot.",
    )
    parser.add_argument(
        "--sleep-seconds",
        type=int,
        default=30,
        help="Seconds to sleep between build status checks.",
    )
    parser.add_argument(
        "--log-max-bytes",
        type=int,
        default=200_000,
        help="Max bytes of log content to attach to instructions.",
    )
    return parser.parse_args(argv)


def _current_branch(repo_path: Path) -> str:
    result = subprocess.run(
        ["git", "rev-parse", "--abbrev-ref", "HEAD"],
        cwd=repo_path,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(f"Failed to resolve git branch: {result.stderr.strip()}")
    return result.stdout.strip()


def main(argv: list[str] | None = None) -> int:
    args = _parse_args(argv or sys.argv[1:])

    repo_path = Path.cwd().resolve()
    if not (repo_path / ".git").exists():
        print(f"Current directory is not a git repo: {repo_path}")
        return 2

    branch = _current_branch(repo_path)
    if branch == "main":
        print("Refusing to run on branch 'main'. Create a working branch first.")
        return 2
    if branch == "HEAD":
        print("Refusing to run on detached HEAD. Create a working branch first.")
        return 2

    ado_pat = args.pat or os.getenv("ADO_PAT", None)

    if not ado_pat:
        print("An ADO PAT must be provided in --pat or ADO_PAT environment variable.")
        return 2

    try:
        ado_org, ado_project, build_def, ado_base_url, build_id = builds.parse_build_url(
            args.build
        )
    except ValueError as exc:
        print(f"Invalid --build URL: {exc}")
        return 2

    return push.run(
        issue=args.issue,
        build_def=build_def,
        start_build_id=int(build_id) if build_id else None,
        repo_path=repo_path,
        llm=args.llm,
        sleep_seconds=args.sleep_seconds,
        log_max_bytes=args.log_max_bytes,
        ado_org=ado_org,
        ado_project=ado_project,
        ado_base_url=ado_base_url,
        pat=ado_pat,
    )


if __name__ == "__main__":
    raise SystemExit(main())
