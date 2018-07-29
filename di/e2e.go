// +build e2e

package di

import "sync"

func Reset() {
	reset(
		&envOnce,
		&jsonCLIPresenterOnce,
		&jsonFileInputterOnce,
		&promptInputterOnce,
		&gRPCClientOnce,
		&dynamicBuilderOnce,
		&initerOnce,
	)
}

func reset(o ...*sync.Once) {
	for _, once := range o {
		*once = sync.Once{}
	}
}
