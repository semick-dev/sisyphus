from pathlib import Path

from agent import payload
from agent.ado import logs


def test_truncate_respects_max_bytes():
    text = "a" * 10
    truncated = logs._truncate(text, max_bytes=5)
    assert len(truncated.encode("utf-8")) <= 5


def test_build_failure_instructions_includes_log(monkeypatch, tmp_path: Path):
    monkeypatch.setattr(logs, "fetch_failure_excerpt", lambda *_args, **_kwargs: "boom")
    class DummyClient:
        pass

    result = payload.build_failure_instructions(
        issue="Org/repo#1",
        build_def="99",
        repo_path=tmp_path,
        build_id=123,
        client=DummyClient(),
        log_max_bytes=10,
    )
    assert "boom" in result
    assert "123" in result
