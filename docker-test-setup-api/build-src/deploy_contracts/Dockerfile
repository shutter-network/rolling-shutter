FROM alpine/git AS builder

RUN date
RUN git clone --branch=shutter-api https://github.com/shutter-network/contracts /contracts
RUN cd /contracts && git submodule update --init --recursive

FROM --platform=linux/amd64 ghcr.io/foundry-rs/foundry AS runner
WORKDIR /contracts
COPY --from=builder --chown=foundry:foundry /contracts /contracts

RUN git config --global --add safe.directory /contracts

USER foundry
RUN forge install
RUN forge build
ENTRYPOINT ["forge", "script"]