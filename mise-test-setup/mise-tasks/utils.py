from __future__ import annotations

import json
import os
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


def load_deployment_run() -> dict[str, object]:
    deployment_type = resolve_deployment_type(os.environ.get("DEPLOYMENT_TYPE", ""))
    data_dir = Path(os.environ["DATA_DIR"])
    run_path = (
        data_dir
        / "contracts"
        / "broadcast"
        / DEPLOYMENT_SCRIPTS[deployment_type]
        / os.environ["ETHEREUM_CHAIN_ID"]
        / "run-latest.json"
    )
    return json.loads(run_path.read_text())


def get_deployed_address(contract_name: str) -> str:
    address = get_created_contract_address(load_deployment_run(), contract_name)
    if not address:
        raise SystemExit(f"{contract_name} address not found in deployment broadcast")
    return address


def cast_call(address: str, selector: str, *args: str) -> str:
    return run(
        [
            "docker", "compose", "run", "--rm", "--entrypoint", "cast",
            "contracts", "call",
            "--rpc-url", "http://ethereum:8545",
            address, selector, *args,
        ],
        capture_output=True,
    ).stdout.strip()


def cast_block_number() -> int:
    return int(
        run(
            [
                "docker", "compose", "run", "--rm", "--entrypoint", "cast",
                "contracts", "block-number",
                "--rpc-url", "http://ethereum:8545",
            ],
            capture_output=True,
        ).stdout.strip()
    )


def get_latest_keyper_set_index() -> int:
    """Return the highest Keyper Set Index registered on-chain.

    Index 0 is the bootstrap guard keyper set; real sets start at 1.
    """
    ksm = get_deployed_address("KeyperSetManager")
    count = int(cast_call(ksm, "getNumKeyperSets()(uint64)"))
    if count == 0:
        raise SystemExit("No keyper sets exist on-chain")
    return count - 1


def get_dkg_start(ksi: int, retry: int) -> int:
    dkg = get_deployed_address("DKGContract")
    return int(cast_call(dkg, "dkgStart(uint64,uint64)(int256)", str(ksi), str(retry)))


def get_cycle_length() -> int:
    dkg = get_deployed_address("DKGContract")
    return int(cast_call(dkg, "cycleLength()(uint64)"))


def get_phase_length() -> int:
    dkg = get_deployed_address("DKGContract")
    return int(cast_call(dkg, "PHASE_LENGTH()(uint64)"))


def cast_logs_at_address(
    address: str,
    *,
    from_block: str = "earliest",
    to_block: str | None = None,
) -> list[dict]:
    """Fetch all logs for the given contract address as parsed JSON entries.

    A single `cast logs` invocation; no topic filtering — callers narrow
    results client-side via topic0 (event signature) and topic1..3.
    """
    cmd = [
        "docker", "compose", "run", "--rm", "--entrypoint", "cast",
        "contracts", "logs",
        "--rpc-url", "http://ethereum:8545",
        "--address", address,
        "--from-block", from_block,
    ]
    if to_block is not None:
        cmd += ["--to-block", to_block]
    cmd.append("--json")
    result = run(cmd, capture_output=True).stdout.strip()
    if not result:
        return []
    return json.loads(result)


def topic_to_uint(topic: str) -> int:
    return int(topic, 16)


def get_succeeded(ksi: int) -> bool:
    dkg = get_deployed_address("DKGContract")
    result = cast_call(dkg, "succeeded(uint64)(bool)", str(ksi))
    return result.lower() == "true"


def derive_retry_counter(ksi: int) -> int:
    """Derive the active retry counter for a Keyper Set Index from the current block.

    Returns the largest n such that dkgStart(ksi, n) <= currentBlock, or 0 if
    no retry has started yet.
    """
    current = cast_block_number()
    start_0 = get_dkg_start(ksi, 0)
    cycle = get_cycle_length()
    if current < start_0:
        return 0
    return (current - start_0) // cycle


def retry_window_elapsed(ksi: int, retry: int) -> bool:
    """Return True once the full cycle window for (ksi, retry) has passed."""
    start = get_dkg_start(ksi, retry)
    cycle = get_cycle_length()
    return cast_block_number() > start + cycle
