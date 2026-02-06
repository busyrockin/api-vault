import json
import os
import tempfile
from datetime import datetime, timezone
from pathlib import Path

LOG_PATH = Path.home() / ".api-vault" / "approvals.json"


class Approvals:
    def log_access(self, credential: str, project: str = "") -> None:
        data = self._read()
        data["log"].append({
            "credential": credential,
            "project": project or os.getcwd(),
            "accessed_at": datetime.now(timezone.utc).isoformat(),
        })
        self._write(data)

    def history(self, credential: str | None = None) -> list[dict]:
        entries = self._read()["log"]
        if credential:
            return [e for e in entries if e["credential"] == credential]
        return entries

    @staticmethod
    def _read() -> dict:
        if LOG_PATH.exists():
            return json.loads(LOG_PATH.read_text())
        return {"log": []}

    @staticmethod
    def _write(data: dict) -> None:
        LOG_PATH.parent.mkdir(parents=True, exist_ok=True)
        tmp = tempfile.NamedTemporaryFile(
            mode="w", dir=LOG_PATH.parent, suffix=".tmp", delete=False,
        )
        try:
            json.dump(data, tmp, indent=2)
            tmp.close()
            os.replace(tmp.name, LOG_PATH)
        except BaseException:
            os.unlink(tmp.name)
            raise
