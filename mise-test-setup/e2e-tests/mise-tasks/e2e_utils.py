"""Shared helpers for DKG e2e tests.

Importable from `mise run` task scripts in this directory: the file is colocated
with the task scripts, so uv adds its directory to sys.path automatically.

Both helpers delegate to `mise run wait-for-dkg`, which polls the DKG Contract
directly. They wrap the subprocess in a timeout so that a hung DKG fails the
e2e test promptly rather than blocking the whole suite.
"""

from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path


_PARENT_TASKS_DIR = Path(__file__).resolve().parents[2] / "mise-tasks"
if str(_PARENT_TASKS_DIR) not in sys.path:
    sys.path.insert(0, str(_PARENT_TASKS_DIR))

import utils  # noqa: E402  (re-exported for test scripts)


DEFAULT_SUCCESS_TIMEOUT = 120.0
DEFAULT_FAILURE_TIMEOUT = 240.0


def wait_for_dkg_success(
    *,
    keyper_set_index: int,
    timeout: float = DEFAULT_SUCCESS_TIMEOUT,
) -> None:
    """Wait for `succeeded(keyper_set_index)` on the DKG Contract."""
    cmd = [
        "mise", "run", "wait-for-dkg",
        "--ksi", str(keyper_set_index),
        "--success",
    ]
    try:
        subprocess.run(cmd, check=True, timeout=timeout)
    except subprocess.TimeoutExpired:
        raise SystemExit(
            f"Timed out after {timeout:.0f}s waiting for DKG success "
            f"for keyper set {keyper_set_index}"
        )


def wait_for_dkg_failure(
    *,
    keyper_set_index: int,
    retry_counter: int = 0,
    timeout: float = DEFAULT_FAILURE_TIMEOUT,
) -> None:
    """Wait for the (keyper_set_index, retry_counter) cycle to elapse without success.

    Raises SystemExit if the DKG unexpectedly succeeded or on timeout.
    """
    cmd = [
        "mise", "run", "wait-for-dkg",
        "--ksi", str(keyper_set_index),
        "--retry", str(retry_counter),
        "--failure",
    ]
    try:
        subprocess.run(cmd, check=True, timeout=timeout)
    except subprocess.TimeoutExpired:
        raise SystemExit(
            f"Timed out after {timeout:.0f}s waiting for DKG failure "
            f"for keyper set {keyper_set_index} retry {retry_counter}"
        )
