import os
import signal

import click
from jsonrpcserver import method, Result, Success, serve
from structlog import get_logger

log = get_logger(__name__)


@method
def shutter_set_proposal_key(proposalId: str, key: str) -> Result:
    log.info("JSONRPC", method='shutter_set_proposal_key', proposal_id=proposalId, key=key)
    return Success(True)


@method
def shutter_set_eon_pubkey(eonId: str, key: str) -> Result:
    log.info("JSONRPC", method='shutter_set_eon_pubkey', eon_id=eonId, key=key)
    return Success(True)


@click.command()
@click.option('-h', '--host', default='0.0.0.0')
@click.option('-p', '--port', default=5000)
def main(host: str, port: int) -> None:
    signal.signal(signal.SIGTERM, lambda sig, frame: os.exit(0))
    signal.signal(signal.SIGINT, lambda sig, frame: os.exit(0))
    log.info("Starting Snapshot dummyserver on %s:%d", host, port)
    serve(host, port)


if __name__ == "__main__":
    main()
