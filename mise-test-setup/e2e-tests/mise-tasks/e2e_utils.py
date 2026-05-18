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

2. `wait_for_dkg_failure` — polls keyper-0's `dkg_result` directly for any
   `success = 'f'` row. Used by the offline-recovery test to assert that DKG
   actually failed below threshold rather than just stalling.

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
DEFAULT_FAILURE_TIMEOUT = 90.0


def _poll_interval() -> float:
    return float(os.environ.get("DKG_RESULT_POLL_INTERVAL", "1"))


def _has_dkg_row(keyper_index: int, success: str) -> bool:
    result = utils.query_keyper_db(
        keyper_index,
        f"SELECT 1 FROM dkg_result WHERE success = '{success}' LIMIT 1",
    )
    return result == "1"


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
        try:
            subprocess.run(
                [
                    "mise",
                    "run",
                    "wait-for-dkg",
                    "--keyper-set-index",
                    str(keyper_set_index),
                ],
                check=True,
                timeout=timeout,
            )
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
    timeout: float = DEFAULT_FAILURE_TIMEOUT,
    keyper_index: int = 0,
) -> None:
    """Wait for any DKG failure row, exiting nonzero on timeout.

    Polls keyper-`keyper_index`'s `dkg_result` for any row with
    `success = 'f'`.
    """
    deadline = time.monotonic() + timeout
    poll_interval = _poll_interval()
    while time.monotonic() < deadline:
        if _has_dkg_row(keyper_index, "f"):
            print(f"DKG failure observed on keyper-{keyper_index}")
            return
        time.sleep(poll_interval)
    raise SystemExit(
        f"Timed out after {timeout:.0f}s waiting for DKG failure "
        f"on keyper-{keyper_index}"
    )
