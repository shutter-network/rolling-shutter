module github.com/shutter-network/shutter/shuttermint

go 1.16

require (
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/Workiva/go-datastructures v1.0.53 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/continuity v0.1.0 // indirect
	github.com/deepmap/oapi-codegen v1.9.0
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgraph-io/ristretto v0.0.3 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/ethereum/go-ethereum v1.10.12
	github.com/gballet/go-libpcsclite v0.0.0-20191108122812-4678299bea08 // indirect
	github.com/getkin/kin-openapi v0.80.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgconn v1.10.1
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.14.0
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/koron/go-ssdp v0.0.2 // indirect
	github.com/kr/pretty v0.3.0
	github.com/kyleconroy/sqlc v1.11.0
	github.com/libp2p/go-libp2p v0.14.4
	github.com/libp2p/go-libp2p-core v0.8.6
	github.com/libp2p/go-libp2p-crypto v0.1.0
	github.com/libp2p/go-libp2p-noise v0.2.2 // indirect
	github.com/libp2p/go-libp2p-peerstore v0.2.10 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.5.5
	github.com/libp2p/go-tcp-transport v0.2.8 // indirect
	github.com/mattn/go-runewidth v0.0.10 // indirect
	github.com/miekg/dns v1.1.43 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20210601165009-122bf33a46e0 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/mapstructure v1.4.2
	github.com/multiformats/go-base32 v0.0.4 // indirect
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/multiformats/go-multihash v0.0.16 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/shirou/gopsutil v3.21.9+incompatible // indirect
	github.com/shutter-network/shutter/shlib v0.1.9
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/status-im/keycard-go v0.0.0-20200402102358-957c09536969 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.14
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tyler-smith/go-bip39 v1.0.2 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/crypto v0.0.0-20211108221036-ceb1ce70b4fa
	golang.org/x/net v0.0.0-20211020060615-d418f374d309 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20211109065445-02f5c0300f6e // indirect
	golang.org/x/tools v0.1.7
	google.golang.org/genproto v0.0.0-20211021150943-2b146023228c // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/ini.v1 v1.62.1 // indirect
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
