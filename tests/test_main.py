import types
from pathlib import Path

import pytest

from agent import main as main_mod


def test_parse_build_url_visualstudio():
    url = "https://sbeddall.visualstudio.com/Investigations/_build?definitionId=1"
    org, project, build_def, base_url, build_id = main_mod.builds.parse_build_url(url)
    assert org == "sbeddall"
    assert project == "Investigations"
    assert build_def == "1"
    assert base_url == "https://dev.azure.com"
    assert build_id is None


def test_parse_build_url_dev_azure():
    url = "https://dev.azure.com/myorg/myproject/_build?definitionId=42"
    org, project, build_def, base_url, build_id = main_mod.builds.parse_build_url(url)
    assert org == "myorg"
    assert project == "myproject"
    assert build_def == "42"
    assert base_url == "https://dev.azure.com"
    assert build_id is None


def test_parse_build_url_results_build_id():
    url = "https://sbeddall.visualstudio.com/Investigations/_build/results?buildId=447&view=results"
    org, project, build_def, base_url, build_id = main_mod.builds.parse_build_url(url)
    assert org == "sbeddall"
    assert project == "Investigations"
    assert build_def == ""
    assert base_url == "https://dev.azure.com"
    assert build_id == "447"


def test_main_refuses_main_branch(monkeypatch, tmp_path):
    repo = tmp_path / "repo"
    repo.mkdir()
    (repo / ".git").mkdir()

    monkeypatch.setattr(main_mod.Path, "cwd", lambda: repo)
    monkeypatch.setattr(main_mod, "_current_branch", lambda _: "main")

    argv = [
        "--issue",
        "Org/repo#1",
        "--build",
        "https://sbeddall.visualstudio.com/Investigations/_build?definitionId=1",
        "--pat",
        "token",
    ]
    assert main_mod.main(argv) == 2


def test_current_branch_error(monkeypatch):
    def fake_run(*_args, **_kwargs):
        return types.SimpleNamespace(returncode=1, stdout="", stderr="boom")

    monkeypatch.setattr(main_mod.subprocess, "run", fake_run)
    with pytest.raises(RuntimeError):
        main_mod._current_branch(Path("."))
