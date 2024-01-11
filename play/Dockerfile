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

WORKDIR /src
