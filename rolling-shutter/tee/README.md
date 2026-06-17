# TEE-fortified Keyper

The Keyper software has been fortified with TEE capabilities (currently only supporting Intel® SGX™), securing the secret key storage using hardware-based encryption.
This raises the confidence level of the non-collusion security assumption of the Keypers, as it makes it harder for attackers to access the key material.
Now, it is less likely that an honest Keyper operator can get compromised by an attacker, ensuring his honest behaviour.

## How it works

**Hardware-based process isolation**&emsp;
TEEs use a hardware-level process isolation mechanism that prevents even the operating system and kernel from inspecting or manipulating the process' memory.
This means that secrets are completely safe from even administrator-level attackers, as long as they reside in process memory (assuming the hardware is not compromised).

**Hardware-based encryption**&emsp;
Additionally, we employ the _sealing_ mechanism which allows us to encrypt and authenticate data using hardware-derived keys.
This mechanism allows us to use encryption keys that can either only be derived by a specific executable, or executables signed by a specific authority.
Using this, we can securely store secrets on the disk without the risk of them being deciphered by an attacker.

### Warning

Create a backup of the old persistence files before trying to run this in production on SGX.
The TEE keyper can load old persistence files, but the new persistence files written by it are sealed and therefore no longer compatible with non-TEE keypers.
It is therefore best to create a backup on cold storage somewhere.

## Roadmap

1. Done: Currently, besides ensuring that it can run inside SGX to benefit from strong process isolation, the Keyper code has been adapted to seal its checkpoints before persisting them to the disk.
2. Coming up: In the next step, we eliminate the trust in the blockchain node that is providing the keyper with information about the outside world. We fully verify the blockchain consensus inside the TEE, and we check the merkle proofs for all blocks to ensure correctness and completeness of the blockchain events. This mainly prevents attacks that try to make the keyper reveal its key shares too early by providing fake blockchain data.
3. Future perspective: An outsider's confidence and trust in a Keyper operator can be improved using further measures such as remote attestation (provable computation), which would allow the Keyper to publicly prove that he is running his node inside a TEE. Further fortifications could then fortify the initial cryptographic setup.

The currently planned fortifications only ensure that the Keyper operator can trust the software running on his machine.
As the Keypers already have access to their keys, if they wanted to, they could use those to misbehave at any time.
Since the keypers cannot provably forget their keys, they cannot prove that their keys are entirely controlled by the TEE.
Compatibility-breaking changes to the Shutter protocol would have to be enacted to further improve trust.

# How to set up Shutter to use SGX

1. Clone the Shutter repository and install Go.
2. Install EGo, following the [official instructions](https://docs.edgeless.systems/ego/getting-started/install) for your OS.
3. Build shutter using `GO=ego-go make build`.


**Testing**&emsp;
To run shutter tests inside SGX, create a compiled test executable:

```sh
ego-go test -c
```

Note that currently, this has to be done for each package separately.

**Running an executable inside the TEE**&emsp;
Then, sign and run the resulting test executable, run

```sh
ego sign package.test
ego run package.test
```

This signs the executable and generates an `enclave.json` file in the current directory, which contains the launch configuration for the TEE enclave.
The resulting executable can also be run normally without SGX by invoking it like a normal executable, and conversely, can also be run on a non-SGX CPU in a TEE-emulation mode by setting `OE_SIMULATION=1` before running it using `ego run`.
This step is the same for test executables, as well as the normal shutter executables.

For maximum security, we recommend that there is a trusted central authority responsible for the `ego sign` step of the keyper executable.
This creates an additional layer of security, because an attacker in possession of the signing key can create executables that can unseal the persistence files.
Therefore, the signing key should either be on cold storage, and each keyper signs for himself, or it should be one official key handled by a trusted party.

## TEE-aware programming

Currently, the TEE-specific functionality is contained within the `keyper/tee` package.
Using `tee.HasTEE()`, code can query whether it is running inside a TEE (but does not differentiate between true hardware support and the emulation mode).
This way, code paths can be introduced that will only activate when the executable has been explicitly invoked as a TEE-capable process using `ego run`, and so the same code can stay compatible with non-TEE hardware.