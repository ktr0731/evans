package entity

type Service interface {
	Name() string
	FQRN() string
	RPCs() []RPC
}
