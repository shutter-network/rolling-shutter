package event

type (
	KeyperSetHandler    func(*KeyperSet) error
	EonPublicKeyHandler func(*EonPublicKey) error
	BlockHandler        func(*LatestBlock) error
	ShutterStateHandler func(*ShutterState) error
)
