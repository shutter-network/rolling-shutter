#!/usr/bin/env bash
set -xe

BB="docker run --rm -v$(pwd)/data:/data -w / busybox"
if docker compose ls >/dev/null 2>&1; then
  # compose v2
  DC="docker compose"
else
  DC=docker-compose
fi

$DC stop geth
$DC rm -f geth
$DC stop chain-{0..2}
$DC rm -f chain-{0..2}

${BB} rm -rf data/geth
${BB} rm -rf data/chain-{0..2}
${BB} mkdir -p data/chain-{0..2}/config
${BB} rm -rf data/deployments

$DC up deploy-contracts  # has geth as dependency


for num in 0 1 2; do
cmd=chain-$num
$DC run --rm --no-deps ${cmd} init \
  --root /chain \
  --genesis-keyper 0x440Dc6F164e9241F04d282215ceF2780cd0B755e \
  --blocktime 1 \
  --listen-address tcp://0.0.0.0:26657
sed -i "/ValidatorPublicKey/c\ValidatorPublicKey = \"$(cat data/chain-${num}/config/priv_validator_pubkey.hex)\"" config/keyper-${num}.toml
if [ $num -eq 0 ];
then
    ${BB} cp data/chain-0/config/genesis.json data/chain-1/config/
    ${BB} cp data/chain-0/config/genesis.json data/chain-2/config/
fi
done

bootstrap_peers=$(cat data/chain-0/config/node_key.json.id)@chain-0:26656
for num in 1 2; do
cmd=chain-$num
bootstrap_peers=${bootstrap_peers},$(cat data/${cmd}/config/node_key.json.id)@${cmd}:26656
done

for num in 0 1 2; do
  cmd=chain-$num
  peers=$(echo ${bootstrap_peers}|cut -d',' -f $(( ((num + 1) % 3) + 1)),$(( ((num + 2) % 3) + 1)))
  ${BB} sed -i "/^persistent-peers =/c\persistent-peers = \"${peers}\"" data/${cmd}/config/config.toml
  done
$DC up -d chain-{0..2} keyper-{0..2}

echo "We need to wait for the chain to reach height >= 1"
sleep 5
echo "This will take a while..."

for num in 0; do
cmd=chain-$num
$DC run --rm --no-deps --entrypoint /rolling-shutter ${cmd} bootstrap \
  --deployment-dir /deployments/dockerGeth \
  --ethereum-url http://geth:8545 \
  --shuttermint-url http://$cmd:26657 \
  --signing-key 479968ffa5ee4c84514a477a8f15f3db0413964fd4c20b08a55fed9fed790fad
done
$DC stop -t 30 geth chain-{0..2} keyper-{0..2}
