FROM alpine/git AS builder

RUN date
RUN git clone --branch=shutter-api https://github.com/shutter-network/contracts /contracts

FROM --platform=linux/amd64 ghcr.io/foundry-rs/foundry AS runner
WORKDIR /contracts
COPY --from=builder /contracts /contracts

RUN forge install
RUN forge build
ENTRYPOINT ["forge", "script"]