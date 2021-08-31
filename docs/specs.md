# Specs

Here, we try to estimate the performance specs of Rolling Shutter, relative to a
plain rollup.

We assume there is demand for 10 transactions per second.

- epoch period: 10 seconds [^1]
- inclusion latency: immediate [^2]
- decryption latency: ~5s [^3]
- confirmation latency: 10 minutes [^4]
- operating costs per tx: 0 [^5]
- batch confirmation mainnet gas per tx: 10 [^6]
- batch inclusion proof mainnet gas per tx: 512 [^7]

<!-- prettier-ignore -->
[^1] In each epoch, the following tasks have to be performed:
    1. collator: close the batch
    2. keypers: generate a decryption key
    3. decryptors: decrypt the batch and sign the result
    4. collator: execute the transactions in the rollup (but not necessarily on
       the mainchain)
    Steps 1 and 4 are local and take an negligible amount of time. Steps 2 and 3
    require one broadcast round among keypers and decryptors respectively, which
    is probably the limiting factor for the batch time. 10s should be plenty of
    time for this. The benefit of going shorter is unclear as users get an
    initial confirmation for their transactions immediately anyway.

<!-- prettier-ignore -->
[^2] The collator can give a preconfirmation that they will include the
    transaction when they close the batch. This promise can be made binding
    via a Merkle proof that allows users to slash the collator if they don't
    abide by it.

<!-- prettier-ignore -->
[^3] The decryption latency depends on the time when the transaction was sent
    relative to the closing time of the batch. If users always send transactions
    for the current batch, it would be 5s on average (half a batch period). In
    practice it might be a little bit higher if users don't send them for the
    current batch at the end of the batching period, but for the next one.

<!-- prettier-ignore -->
[^4] Depends on how often the sequencer sends mainchain transactions.

<!-- prettier-ignore -->
[^5] Costs of running nodes. Assuming 1000 nodes costing $10 per month and
    10tx/s this results in `1000 * 10 $/month / (10 tx/s * 60 * 60 * 24 * 28 s/month) = 0.0004 $`.

<!-- prettier-ignore -->
[^6] Cost of verifying an aggregate BLS signature in the rollup, plus submitting
    the batch hash. On mainnet, this is just the cost of including the calldata
    (16 gas/byte). The signature is 64 bytes, the hash 32, so at 100tx/batch the
    cost per transaction is about 15 gas.

<!-- prettier-ignore -->
[^7] Each transaction needs to prove that it is part of the batch. To do that,
    it has to provide at least one hash of 32 bytes, resulting in 512 mainnet
    gas costs. Maybe multiple transactions can share this cost, but only if they
    are included as one.
