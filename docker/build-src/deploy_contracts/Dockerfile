FROM node:16.16

RUN mkdir /contracts

COPY /contracts/package*.json /contracts/
WORKDIR /contracts

RUN --mount=type=cache,target=/root/.npm npm install

COPY /contracts /contracts

ENTRYPOINT ["npx", "hardhat", "--network", "dockerGeth"]
CMD ["deploy"]
