FROM node:16.16

COPY /contracts /contracts
WORKDIR /contracts

RUN npm install

HEALTHCHECK CMD curl --fail http://localhost:8545
ENTRYPOINT ["npx", "hardhat"]
CMD ["node"]
