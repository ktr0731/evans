package config

import (
	"github.com/ktr0731/evans/logger"
	"github.com/spf13/viper"
)

var migrationScripts = map[string]func(string, *viper.Viper) string{
	"0.6.10": migrate0610To0611,
}

// migrate migrates an old config schema to the latest one.
// migrate accepts an old version and its config. Migration
// flow is as follows:
//
//   1. check whether a migration script which migrates the
//      current version to another version. If it is found,
//      apply it to the current config v. Each script is
//      managed by a variable migrationScripts.
//
//   2. The migration script returns a version which is the
//      same as an updated version.
//
//   3. migrate instructs 1 and 2 again with the returned
//      version. If there isn't a new migration script,
//      migrate finishes migration processing.
//
func migrate(old string, v *viper.Viper) {
	f, ok := migrationScripts[old]
	if !ok {
		// No any changes
		return
	}
	updatedVer := f(old, v)
	migrate(updatedVer, v)
}

// migrate0610To0611 migrates a v0.6.10 or older config to v0.6.11 config.
func migrate0610To0611(old string, v *viper.Viper) string {
	const updatedVer = "0.6.11"

	v.Set("meta.configVersion", updatedVer)

	oldHeader := []*struct {
		Key string `toml:"key"`
		Val string `toml:"val"`
	}{}

	// request.header in v0.6.10 is formatted like:
	//
	//   [[request.header]]
	//     key = "grpc-client"
	//     val = "evans"
	//
	// These are parsed as a map like:
	//
	//   []interface {}{
	//     {
	//       "key": "grpc-client",                                                                                                                          "val": "evans",
	//       "val": "evans",
	//     },
	//   }
	//
	if err := v.UnmarshalKey("request.header", &oldHeader); err != nil {
		logger.Printf("failed to unmarshal 'request.header' in v%s", old)
		return ""
	}

	// v0.6.11 modifies the above structure to a map.
	m := make(map[string]string)
	for _, h := range oldHeader {
		m[h.Key] = h.Val
	}
	v.Set("request.header", m)

	// v0.6.11 renamed Input.PromptFormat to REPL.InputPromptFormat.
	v.Set("repl.inputPromptFormat", v.Get("input.promptFormat"))
	// v0.6.11 removed Input field.
	v.Set("input", nil)

	return updatedVer
}
