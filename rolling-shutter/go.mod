module github.com/shutter-network/shutter/shuttermint

go 1.16

require (
	github.com/deepmap/oapi-codegen v1.9.0
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgraph-io/ristretto v0.0.3 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/ethereum/go-ethereum v1.10.16
	github.com/fjl/memsize v0.0.1 // indirect
	github.com/gballet/go-libpcsclite v0.0.0-20191108122812-4678299bea08 // indirect
	github.com/getkin/kin-openapi v0.80.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/go-bexpr v0.1.11 // indirect
	github.com/influxdata/influxdb v1.8.10 // indirect
	github.com/influxdata/line-protocol v0.0.0-20210922203350-b1ad95c89adf // indirect
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgconn v1.10.1
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.14.1
	github.com/jackc/puddle v1.2.1 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/koron/go-ssdp v0.0.2 // indirect
	github.com/kr/pretty v0.3.0
	github.com/kyleconroy/sqlc v1.11.0
	github.com/labstack/echo/v4 v4.2.2 // indirect
	github.com/labstack/gommon v0.3.1 // indirect
	github.com/libp2p/go-libp2p v0.14.4
	github.com/libp2p/go-libp2p-core v0.8.6
	github.com/libp2p/go-libp2p-noise v0.2.2 // indirect
	github.com/libp2p/go-libp2p-peerstore v0.2.10 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.5.6
	github.com/libp2p/go-tcp-transport v0.2.8 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20210601165009-122bf33a46e0 // indirect
	github.com/mitchellh/mapstructure v1.4.3
	github.com/multiformats/go-base32 v0.0.4 // indirect
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/multiformats/go-multihash v0.0.16 // indirect
	github.com/pingcap/log v0.0.0-20211207084639-71a2e5860834 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.30.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shutter-network/shutter/shlib v0.1.9
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/status-im/keycard-go v0.0.0-20211109104530-b0e0482ba91d // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.15
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tyler-smith/go-bip39 v1.0.2 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/net v0.0.0-20211206223403-eba003a116a9 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	golang.org/x/tools v0.1.8
	google.golang.org/genproto v0.0.0-20211206220100-3cb06788ce7f // indirect
	google.golang.org/protobuf v1.27.1
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.3
)

// The exclude directive for tendermint/tm-db is needed because this
// version is incompatible with tendermint 0.34.13 and it prevents us
// from running 'go get -u=patch':
// ,----
// | % go get -u=patch
// | # github.com/tendermint/tendermint/abci/example/kvstore
// | ../../../go/pkg/mod/github.com/tendermint/tendermint@v0.34.13/abci/example/kvstore/kvstore.go:74:21: undefined: db.NewMemDB
// | ../../../go/pkg/mod/github.com/tendermint/tendermint@v0.34.13/abci/example/kvstore/persistent_kvstore.go:40:13: undefined: db.NewGoLevelDB
// `----
exclude github.com/tendermint/tm-db v0.6.5
