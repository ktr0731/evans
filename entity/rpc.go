package entity

type RPC interface {
	Name() string
	FQRN() string
	RequestMessage() Message
	ResponseMessage() Message
}
