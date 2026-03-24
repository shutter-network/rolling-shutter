from __future__ import annotations

import subprocess
import time
from pathlib import Path


DEPLOYMENT_SCRIPTS: dict[str, str] = {
    "gnosis": "Deploy.gnosh.s.sol",
    "service": "Deploy.service.s.sol",
}

KEYPER_SUBCOMMANDS: dict[str, str] = {
    "gnosis": "gnosiskeyper",
    "service": "shutterservicekeyper",
}


def run(
    command: list[str], *, capture_output: bool = False
) -> subprocess.CompletedProcess[str]:
    return subprocess.run(command, check=True, text=True, capture_output=capture_output)


def wait_for_service_health(service: str, *, timeout_seconds: float = 30.0) -> None:
    container_id = run(
        ["docker", "compose", "ps", "-q", service], capture_output=True
    ).stdout.strip()
    if not container_id:
        raise SystemExit(f"Missing container for {service}")

    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        status = run(
            [
                "docker",
                "inspect",
                "-f",
                "{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}",
                container_id,
            ],
            capture_output=True,
        ).stdout.strip()
        if status == "healthy":
            return
        if status == "unhealthy":
            raise SystemExit(f"Service {service} became unhealthy")
        if status == "none":
            raise SystemExit(f"Service {service} has no healthcheck")
        time.sleep(0.1)
    raise SystemExit(f"Timed out waiting for {service} to become healthy")


def set_toml_path(document, parts: list[str], value) -> None:
    import tomlkit

    current = document
    for part in parts[:-1]:
        if part not in current or not isinstance(current[part], dict):
            current[part] = tomlkit.table()
        current = current[part]
    current[parts[-1]] = value


def keyper_address(config_path: Path) -> str:
    return config_path.read_text().splitlines()[0].removeprefix("# Ethereum address: ").strip()


def parse_indices(indices: str) -> list[int]:
    return [int(index.strip()) for index in indices.split(",") if index.strip()]


def query_keyper_db(keyper_index: int, sql: str) -> str:
    return run(
        [
            "docker",
            "compose",
            "exec",
            "-T",
            "db",
            "psql",
            "-U",
            "postgres",
            "-d",
            f"keyper-{keyper_index}",
            "-tAc",
            sql,
        ],
        capture_output=True,
    ).stdout.strip()


def resolve_deployment_type(deployment_type: str) -> str:
    if deployment_type in DEPLOYMENT_SCRIPTS:
        return deployment_type
    if deployment_type:
        raise SystemExit(f"Unsupported DEPLOYMENT_TYPE: {deployment_type}")
    raise SystemExit("DEPLOYMENT_TYPE is empty")


def get_created_contract_address(
    deployment_run: dict[str, object], contract_name: str
) -> str | None:
    transactions = deployment_run.get("transactions")
    if not isinstance(transactions, list):
        return None
    for tx in transactions:
        if not isinstance(tx, dict):
            continue
        if tx.get("transactionType") != "CREATE":
            continue
        if tx.get("contractName") != contract_name:
            continue
        address = tx.get("contractAddress")
        if isinstance(address, str) and address:
            return address
    return None
