FROM alpine/git AS builder

RUN date
RUN git clone --branch=docker_deploy_gnosis https://github.com/shutter-network/contracts /contracts

COPY /contracts/package*.json /contracts/
WORKDIR /contracts

#COPY /contracts /contracts

FROM ghcr.io/foundry-rs/foundry:latest AS runner
WORKDIR /contracts
COPY --from=builder /contracts /contracts

RUN forge build
ENTRYPOINT ["forge", "script"]
