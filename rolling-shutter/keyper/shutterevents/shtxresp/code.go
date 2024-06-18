// Used for the return codes between shuttermint and tx senders (e.g. Keyper, etc.)
package shtxresp

const (
	Ok uint32 = iota
	Error
	Seen
)
