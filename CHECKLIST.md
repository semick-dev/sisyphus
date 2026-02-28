# Sisypnus Agent Checklist

- [x] Create `pyproject.toml` with build metadata and console entrypoint.
- [x] Add package scaffold under `agent/`.
- [x] Implement `agent/main.py` entrypoint with `argparse` and validation.
- [x] Implement `agent/push.py` loop orchestration.
- [x] Implement ADO client package under `agent/ado/` (auth, client, builds, logs).
- [x] Implement `agent/payload.py` to build and update `instructions.md`.
- [x] Implement `agent/man.py` ASCII art renderer.
- [x] Add minimal tests (instruction builder + log truncation).
- [ ] Add README updates if needed.
