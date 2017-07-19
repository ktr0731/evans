package repl

import (
	"github.com/lycoris0731/evans/env"
)

func call(env *env.Env, name string) (string, error) {
	// var svcName, rpcName string
	// if env.currentService == "" {
	// 	splitted := strings.Split(name, ".")
	// 	if len(splitted) < 2 {
	// 		return "", errors.Wrap(ErrArgumentRequired, "service or RPC name")
	// 	}
	// 	svcName, rpcName = splitted[0], splitted[1]
	// } else {
	// 	svcName = env.state.currentService
	// 	rpcName = name
	// }

	// 1. 引数、戻り値取得
	// 2. 各フィールド取得
	// 3. 新しいプロンプトを起動
	// rpc, err := env.Desc.GetRPC(svcName, rpcName)
	// if err != nil {
	// 	return "", err
	// }
	// TODO: Type はパッケージ名を取り除く (.test.HelloRequest => HelloRequest)
	// resType := env.Desc.GetMessage(env.state.currentPackage, rpc.RequestType)
	// pp.Println(resType)

	return "", nil
}
