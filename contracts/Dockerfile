FROM node:18.17.1 as builder
WORKDIR /contracts
COPY /contracts /contracts
RUN npm install

FROM gcr.io/distroless/nodejs18-debian11
COPY --from=builder /contracts /contracts
WORKDIR /contracts
# CMD ["server.js"]
