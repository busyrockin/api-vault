import os
import re
import subprocess
from pathlib import Path


class VaultError(Exception):
    pass


class Vault:
    def __init__(self):
        self.binary = self._find_binary()
        if not os.environ.get("API_VAULT_PASSWORD"):
            raise VaultError("API_VAULT_PASSWORD not set")

    def get(self, name: str) -> str:
        return self._run("get", name)

    def list(self) -> list[dict]:
        raw = self._run("list")
        if not raw.strip():
            return []
        lines = raw.strip().splitlines()
        if len(lines) < 2:
            return []
        rows = []
        for line in lines[1:]:
            cols = re.split(r"\s{2,}", line.strip())
            if len(cols) >= 3:
                rows.append({"name": cols[0], "type": cols[1], "created": cols[2]})
        return rows

    def _run(self, *args: str) -> str:
        env = {**os.environ, "API_VAULT_PASSWORD": os.environ["API_VAULT_PASSWORD"]}
        try:
            r = subprocess.run(
                [self.binary, *args],
                capture_output=True, text=True, timeout=10, env=env,
            )
        except subprocess.TimeoutExpired:
            raise VaultError(f"vault command timed out: {args}")
        if r.returncode != 0:
            raise VaultError(r.stderr.strip() or f"vault exited {r.returncode}")
        return r.stdout

    @staticmethod
    def _find_binary() -> str:
        project = Path(__file__).resolve().parent.parent / "api-vault"
        if project.is_file():
            return str(project)
        from shutil import which
        found = which("api-vault")
        if found:
            return found
        raise VaultError("api-vault binary not found — run 'make build'")
