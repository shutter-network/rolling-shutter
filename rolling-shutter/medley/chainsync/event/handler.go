package event

import "context"

type (
	KeyperSetHandler    func(context.Context, *KeyperSet) error
	EonPublicKeyHandler func(context.Context, *EonPublicKey) error
	BlockHandler        func(context.Context, *LatestBlock) error
	ShutterStateHandler func(context.Context, *ShutterState) error
	ECIESKeyHandler     func(context.Context, *ECIESKey) error
)
