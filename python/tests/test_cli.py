"""CLI tests: the README invocation runs end to end and gates on breaking changes."""

import json
import subprocess
import sys
from pathlib import Path

PYTHON_DIR = Path(__file__).resolve().parent.parent


def _write(path: Path, fields: list[dict]) -> str:
    path.write_text(json.dumps({"fields": fields}), encoding="utf-8")
    return str(path)


def _run(*args: str, cwd: Path) -> subprocess.CompletedProcess:
    return subprocess.run(
        [sys.executable, *args], check=False, cwd=cwd, capture_output=True, text=True
    )


def test_readme_command_fails_on_breaking_change(tmp_path):
    old = _write(tmp_path / "old.json", [{"name": "id", "type": "int"}])
    new = _write(
        tmp_path / "new.json",
        [
            {"name": "id", "type": "int"},
            {"name": "email", "type": "string"},
            {"name": "nickname", "type": "string", "required": False},
        ],
    )
    result = _run("-m", "src.cli", old, new, cwd=PYTHON_DIR)
    assert result.returncode == 1, result.stderr
    assert "[BREAKING] email: added required field" in result.stdout
    assert "[safe] nickname: added optional field" in result.stdout
    assert "FAIL: breaking schema changes detected" in result.stdout


def test_readme_command_passes_on_safe_change(tmp_path):
    old = _write(tmp_path / "old.json", [{"name": "id", "type": "int"}])
    new = _write(
        tmp_path / "new.json",
        [{"name": "id", "type": "int"}, {"name": "nickname", "type": "string", "required": False}],
    )
    result = _run("-m", "src.cli", old, new, cwd=PYTHON_DIR)
    assert result.returncode == 0, result.stderr
    assert "OK: all changes are backward-compatible" in result.stdout


def test_cli_runs_from_src_dir(tmp_path):
    old = _write(tmp_path / "old.json", [{"name": "id", "type": "int"}])
    result = _run("-m", "cli", old, old, cwd=PYTHON_DIR / "src")
    assert result.returncode == 0, result.stderr
    assert "no schema changes" in result.stdout
