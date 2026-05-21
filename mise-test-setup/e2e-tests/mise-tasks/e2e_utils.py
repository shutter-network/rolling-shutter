"""Shared helpers for DKG e2e tests.

Importable from `mise run` task scripts in this directory: the file is colocated
with the task scripts, so uv adds its directory to sys.path automatically.

The helpers wrap two ways of asserting DKG outcomes:

1. `wait_for_dkg_success` — for the happy path with multiple keyper set
   transitions, pass `keyper_set_index=N` to delegate to the existing
   `wait-for-dkg --keyper-set-index N` task (which is eon-aware and polls
   through failure rows for prior eons of the same set). Without that argument,
   it polls keyper-0's `dkg_result` directly for any `success = 't'` row — the
   simpler "did DKG ever succeed?" check used by the offline-recovery test
   after triggering a retry.

2. `wait_for_dkg_failure` — polls the DKG contract on-chain. Once the full
   cycle for retry 0 has elapsed without `succeeded(ksi)` returning true, the
   DKG definitively failed. Accepts a required `keyper_set_index` parameter.

Both functions raise `SystemExit` with a clear message on timeout so that a
`mise run` task exits nonzero immediately.
"""

from __future__ import annotations

import os
import subprocess
import sys
import time
from pathlib import Path


_PARENT_TASKS_DIR = Path(__file__).resolve().parents[2] / "mise-tasks"
if str(_PARENT_TASKS_DIR) not in sys.path:
    sys.path.insert(0, str(_PARENT_TASKS_DIR))

import utils  # noqa: E402  (re-exported for test scripts)


DEFAULT_SUCCESS_TIMEOUT = 120.0
DEFAULT_FAILURE_TIMEOUT = 240.0


def _poll_interval() -> float:
    return float(os.environ.get("DKG_RESULT_POLL_INTERVAL", "1"))


def _has_dkg_row(keyper_index: int, success: str) -> bool:
    result = utils.query_keyper_db(
        keyper_index,
        f"SELECT 1 FROM dkg_result WHERE success = '{success}' LIMIT 1",
    )
    return result == "1"


def _cast_call(address: str, selector: str, *args: str) -> str:
    return utils.run(
        [
            "docker", "compose", "run", "--rm", "--entrypoint", "cast",
            "contracts", "call",
            "--rpc-url", "http://ethereum:8545",
            address, selector, *args,
        ],
        capture_output=True,
    ).stdout.strip()


def _cast_block_number() -> int:
    return int(utils.run(
        [
            "docker", "compose", "run", "--rm", "--entrypoint", "cast",
            "contracts", "block-number",
            "--rpc-url", "http://ethereum:8545",
        ],
        capture_output=True,
    ).stdout.strip())


def wait_for_dkg_success(
    *,
    keyper_set_index: int | None = None,
    timeout: float = DEFAULT_SUCCESS_TIMEOUT,
    keyper_index: int = 0,
) -> None:
    """Wait for DKG to succeed, exiting nonzero on timeout.

    With `keyper_set_index`: delegates to `mise run wait-for-dkg
    --keyper-set-index N` (eon-aware, polls through prior failures of the same
    set) and enforces a subprocess-level timeout.

    Without `keyper_set_index`: polls keyper-`keyper_index`'s `dkg_result`
    directly for any row with `success = 't'`.
    """
    if keyper_set_index is not None:
        cmd = [
            "mise", "run", "wait-for-dkg",
            "--keyper-set-index", str(keyper_set_index),
            "--keyper-index", str(keyper_index),
        ]
        try:
            subprocess.run(cmd, check=True, timeout=timeout)
            return
        except subprocess.TimeoutExpired:
            raise SystemExit(
                f"Timed out after {timeout:.0f}s waiting for DKG success "
                f"for keyper set {keyper_set_index}"
            )

    deadline = time.monotonic() + timeout
    poll_interval = _poll_interval()
    while time.monotonic() < deadline:
        if _has_dkg_row(keyper_index, "t"):
            print(f"DKG success observed on keyper-{keyper_index}")
            return
        time.sleep(poll_interval)
    raise SystemExit(
        f"Timed out after {timeout:.0f}s waiting for DKG success "
        f"on keyper-{keyper_index}"
    )


def wait_for_dkg_failure(
    *,
    keyper_set_index: int,
    retry_counter: int = 0,
    timeout: float = DEFAULT_FAILURE_TIMEOUT,
    keyper_index: int = 0,
) -> None:
    """Wait for DKG failure by polling on-chain state.

    Waits until the full DKG cycle for `retry_counter` has elapsed without
    `succeeded(keyper_set_index)` returning true on the DKG contract.
    Raises SystemExit if the DKG unexpectedly succeeded or on timeout.
    """
    deadline = time.monotonic() + timeout
    poll_interval = _poll_interval()

    # Poll until the keypers have observed the activation block and written
    # the eons row (which includes the dkg_contract address).
    dkg_contract = ""
    while time.monotonic() < deadline:
        dkg_contract = utils.query_keyper_db(
            keyper_index,
            f"SELECT dkg_contract FROM eons "
            f"WHERE keyper_config_index = {keyper_set_index} LIMIT 1",
        )
        if dkg_contract:
            break
        time.sleep(poll_interval)
    if not dkg_contract:
        raise SystemExit(
            f"Timed out after {timeout:.0f}s waiting for eons entry "
            f"for keyper set {keyper_set_index}"
        )

    dkg_start = int(_cast_call(
        dkg_contract,
        "dkgStart(uint64,uint64)(int256)",
        str(keyper_set_index),
        str(retry_counter),
    ))
    cycle_length = int(_cast_call(dkg_contract, "cycleLength()(uint64)"))
    failure_block = dkg_start + cycle_length

    while time.monotonic() < deadline:
        if _cast_block_number() > failure_block:
            break
        time.sleep(poll_interval)
    else:
        raise SystemExit(
            f"Timed out after {timeout:.0f}s waiting for DKG failure "
            f"for keyper set {keyper_set_index} "
            f"(need block > {failure_block})"
        )

    succeeded = _cast_call(
        dkg_contract,
        "succeeded(uint64)(bool)",
        str(keyper_set_index),
    )
    if succeeded.lower() == "true":
        raise SystemExit(
            f"Unexpected DKG success for keyper set {keyper_set_index}"
        )
    print(
        f"DKG failure confirmed for keyper set {keyper_set_index} "
        f"(block > {failure_block}, succeeded=false)"
    )
