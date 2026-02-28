from __future__ import annotations

import base64
from dataclasses import dataclass
from typing import Any

import requests
@dataclass
class ADOClient:
    org: str
    project: str
    base_url: str
    pat: str

    def _auth_header(self) -> str:
        token = f":{self.pat}".encode("utf-8")
        b64 = base64.b64encode(token).decode("utf-8")
        return f"Basic {b64}"

    def request(
        self,
        method: str,
        path: str,
        *,
        params: dict | None = None,
        json: dict | None = None,
    ) -> Any:
        url = f"{self.base_url}/{self.org}/{self.project}/{path.lstrip('/')}"
        headers = {
            "Authorization": self._auth_header(),
            "Content-Type": "application/json",
        }
        resp = requests.request(method, url, params=params, json=json, headers=headers, timeout=30)
        resp.raise_for_status()
        if resp.headers.get("Content-Type", "").startswith("application/json"):
            return resp.json()
        return resp.text
