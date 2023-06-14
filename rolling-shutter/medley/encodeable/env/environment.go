package env

//go:generate go-enum --marshal

// ENUM(production, staging, local).
type Environment int

func (x *Environment) Equal(b *Environment) bool {
	return x == b
}
