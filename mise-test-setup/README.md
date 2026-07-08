# Mise Test Setup

`mise test setup` is a local mise-driven setup for running the shutter service
flow with Ethereum, keypers, and supporting infrastructure.

For the normal happy path, the two main commands are:

```bash
mise run wait-for-initial-dkg
mise run test-decryption
```

## Main Tasks

- `mise run wait-for-initial-dkg`

  - Bring the system up, add the initial keyper set, and wait until the initial
    DKG succeeds.

- `mise run add-keyper-set --indices 0,1,2,3 --threshold 3`

  - Add a new keyper set on-chain for the selected keypers.

- `mise run wait-for-dkg --ksi 2 --success`

  - Wait until the DKG for a given Keyper Set Index succeeds on-chain. Without
    `--ksi`, defaults to the latest registered keyper set. Without `--success` /
    `--failure`, exits 0 on any completion. With `--retry N`, pins the watch to
    a specific retry counter; with `--success`/`--failure`, asserts that retry's
    outcome.

- `mise run test-decryption`

  - Submit a decryption trigger with default values and wait for the
    corresponding decryption key.

- `mise run submit-identity-registration`

  - Submit an identity registration transaction. By default it chooses the
    latest eon, a random identity prefix, and a near-future timestamp.

- `mise run wait-for-decryption-key --eon 1 --identity-prefix 0x...`

  - Wait until the decryption key for the given `(eon, identity-prefix)` becomes
    available.

- `mise run clean`
  - Stop the containers and reset the environment.

## Supporting Tasks

These are setup and intermediate tasks that the main flow uses internally. You
can still run them directly if you want to test specific parts of the system:

- `gen-compose`
- `up-db`
- `up-ethereum`
- `deploy`
- `gen-keyper-configs`
- `init-keyper-dbs`
- `up`
- `down`
- `clean`
- `add-initial-keyper-set`

Dependency flow for `wait-for-initial-dkg`:

```text
- `wait-for-initial-dkg`
  - `up`
    - `init-keyper-dbs`
      - `up-db`
        - `gen-compose`
      - `gen-keyper-configs`
        - `deploy`
          - `up-ethereum`
            - `gen-compose`
  - `add-initial-keyper-set`
    - `gen-keyper-configs`
      - `deploy`
        - `up-ethereum`
          - `gen-compose`
```

## Relevant Env Vars

- `DEPLOYMENT_TYPE`

  - Selects the deployment mode (`service` or `gnosis`). Only `service` is fully
    supported.

- `NUM_KEYPERS`

  - Total number of keypers and chain nodes to initialize and run.

- `INITIAL_KEYPER_SET_INDICES`

  - Comma-separated list of keyper indices that belong to the initial
    validator/keyper set.

- `INITIAL_THRESHOLD`

  - Threshold for the initial keyper set.

- `ACTIVATION_DELTA`

  - Activation delay used when adding a keyper set on-chain.

- `DECRYPTION_TRIGGER_TIMESTAMP_DELTA`

  - Default offset used when submitting an identity registration without an
    explicit timestamp.

- `DEPLOY_KEY`
  - Private key used for deployments and contract interactions.

## Task Model

- Tasks use `#MISE depends=[...]` to express ordering.
- Tasks are implemented in bash or Python.
- Tasks are meant to be idempotent.

## Gnosis Note

Support for the gnosis keyper flow is not yet fully implemented in the
`mise test setup`.

- The fully exercised path is `DEPLOYMENT_TYPE=service`.
- Some higher-level tasks, especially around identity registration and
  decryption testing, currently assume the service flow.
