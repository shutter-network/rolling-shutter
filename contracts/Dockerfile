FROM node:18.17.1 as builder
WORKDIR /contracts
COPY /contracts/package.json /contracts/package-lock.json /contracts/
RUN npm install
RUN mkdir -p /contracts/deployments

FROM gcr.io/distroless/nodejs18-debian11
VOLUME /contracts/deployments
COPY /contracts /contracts
COPY --from=builder /contracts /contracts
WORKDIR /contracts
ENTRYPOINT ["/nodejs/bin/node", "node_modules/.bin/hardhat"]
