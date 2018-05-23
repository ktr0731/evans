package entity

import "google.golang.org/grpc"

type RPC interface {
	Name() string
	FQRN() string
	RequestMessage() Message
	ResponseMessage() Message
	IsServerStreaming() bool
	IsClientStreaming() bool

	// StreamDesc returns *grpc.StreamDesc.
	// if both of IsServerStreaming and IsClientStreaming are nil,
	// StreamDesc will panic.
	StreamDesc() *grpc.StreamDesc
}
