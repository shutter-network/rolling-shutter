FROM ubuntu:22.04 AS base

SHELL ["/bin/bash", "-c"]
ENV ASDF_DIR=/home/app/.asdf
ENV BASH_ENV=/home/app/.env

RUN groupadd -g 1000 app && \
    useradd -g 1000 -u 1000 -m app

RUN --mount=type=cache,target=/var/cache/apt \
    apt update && \
    apt install -y git curl

USER 1000
RUN git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.13.1

RUN echo '. $HOME/.asdf/asdf.sh' >> ~/.env && \
    echo '. $HOME/.asdf/completions/asdf.bash' >> ~/.env && \
    echo '. ~/.env' >> ~/.bashrc


FROM base AS build

USER root

RUN --mount=type=cache,target=/var/cache/apt \
    apt install -y build-essential libreadline-dev zlib1g-dev libssl-dev libuuid1 uuid-dev unzip

COPY ./ /src

RUN chown -R app:app /src
USER 1000

WORKDIR /src/rolling-shutter

RUN --mount=type=cache,target=/home/app/.asdf/downloads,uid=1000,gid=1000 \
    --mount=type=cache,target=/home/app/.asdf/tmp,uid=1000,gid=1000 \
    make install-asdf

RUN --mount=type=cache,target=/home/app/.asdf/downloads,uid=1000,gid=1000 \
    --mount=type=cache,target=/home/app/.asdf/tmp,uid=1000,gid=1000 \
    --mount=type=cache,target=/home/app/.cache/go-build,uid=1000,gid=1000 \
    make install-asdf && \
    make install-tools && \
    make build

ENV POSTGRES_DB="testdb"
ENV POSTGRES_HOST_AUTH_METHOD="trust"
ENV POSTGRES_PASSWORD="password"
ENV POSTGRES_USER="app"


RUN sh -c "$(curl --location https://raw.githubusercontent.com/F1bonacc1/process-compose/main/scripts/get-pc.sh)" -- -d -b /home/app/.asdf/bin
USER root
RUN ln -s /tmp/.s.PGSQL.5432 /var/run/.s.pgsql.5432
USER 1000
RUN bash -c "mkdir -p /src/play/work/db && /home/app/.asdf/shims/initdb /src/play/work/db"
WORKDIR /src/play

RUN --mount=type=cache,target=/home/app/.cache,uid=1000,gid=1000 \
    bb --config /src/play/bb.edn build-all && \
    clojure -T:build

ENV ROLLING_SHUTTER_ROOT=/src
ENV ROLLING_SHUTTER_BOOTSTRAP_SIGNING_KEY=479968ffa5ee4c84514a477a8f15f3db0413964fd4c20b08a55fed9fed790fad
ENV ROLLING_SHUTTER_CHAIN_GENESIS_KEYPER=0x440Dc6F164e9241F04d282215ceF2780cd0B755e
USER root
RUN mkdir -p /home/app/.cache/hardhat-nodejs && chown -R 1000:1000 /home/app/.cache/hardhat-nodejs
RUN ln -s /src/contracts/deployments /src/play/deployments

USER 1000
