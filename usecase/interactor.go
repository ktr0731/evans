package usecase

import "github.com/ktr0731/evans/usecase/port"

type Interactor struct {
}

func (i *Interactor) Call(params *port.CallParams) (*port.CallResponse, error) {
	return nil, nil
}
