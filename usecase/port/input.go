package port

type InputPort interface {
	Call(*CallParams) (*CallResponse, error)
}

type CallParams struct {
	SvcName string
	RPCName string
}

type CallResponse struct{}
