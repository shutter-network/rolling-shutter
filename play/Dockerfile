FROM ubuntu:22.04 AS base

SHELL ["/bin/bash", "-c"]
ENV ASDF_DIR=/home/app/.asdf
ENV BASH_ENV=/home/app/.env

RUN groupadd -g 1000 app && \
    useradd -g 1000 -u 1000 -m app

RUN apt update && \
    apt install -y git curl

USER 1000
RUN git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.13.1

RUN echo '. $HOME/.asdf/asdf.sh' >> ~/.env && \
    echo '. $HOME/.asdf/completions/asdf.bash' >> ~/.env && \
    echo '. ~/.env' >> ~/.bashrc


FROM base AS build

USER root

RUN apt install -y build-essential libreadline-dev zlib1g-dev libssl-dev libuuid1 uuid-dev unzip

COPY ./ /src

RUN chown -R app:app /src
USER 1000

WORKDIR /src/rolling-shutter

RUN make install-asdf && \
    make install-tools && \
    make build

RUN go clean -cache -modcache

FROM base AS run

USER root
RUN apt install make

USER 1000
RUN mkdir /home/app/.asdf/installs
COPY --from=build /src /src
COPY --from=build /home/app/.asdf/plugins/ /home/app/.asdf/plugins/
COPY --from=build /home/app/.asdf/installs/babashka/ /home/app/.asdf/installs/babashka/
COPY --from=build /home/app/.asdf/installs/clojure/ /home/app/.asdf/installs/clojure/
COPY --from=build /home/app/.asdf/installs/golang/ /home/app/.asdf/installs/golang/
COPY --from=build /home/app/.asdf/installs/java/ /home/app/.asdf/installs/java/
COPY --from=build /home/app/.asdf/installs/nodejs/ /home/app/.asdf/installs/nodejs/
COPY --from=build /home/app/.asdf/installs/postgres/ /home/app/.asdf/installs/postgres/
COPY --from=build /home/app/.asdf/installs/solidity/ /home/app/.asdf/installs/solidity/
RUN asdf reshim


WORKDIR /src
