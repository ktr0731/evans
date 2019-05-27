package e2e_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ktr0731/evans/logger"
	"github.com/ktr0731/grpc-test/server"
	"github.com/phayes/freeport"
	"go.uber.org/goleak"

	_ "github.com/ktr0731/evans/e2e/statik"
)

// TestMain prepares the test environment for E2E testing. TestMain do following things for clean up the environment.
//
//   - Set log output to ioutil.Discard.
//   - Remove .evans.toml in this project root.
//   - Change $XDG_CONFIG_HOME and $XDG_CACHE_HOME to ignore the root config and cache data.
//     These envvars are reset at the end of E2E testing.
//
func TestMain(m *testing.M) {
	logger.SetOutput(ioutil.Discard)

	b, err := exec.Command("git", "rev-parse", "--show-cdup").Output()
	if err != nil {
		panic(fmt.Sprintf("failed to execute 'git rev-parse --show-cdup'"))
	}
	projRoot := strings.TrimSpace(string(b))
	os.Remove(filepath.Join(projRoot, ".evans.toml"))

	setEnv := func(k, v string) func() {
		old := os.Getenv(k)
		os.Setenv(k, v)
		return func() {
			os.Setenv(k, old)
		}
	}

	configDir := os.TempDir()
	cleanup1 := setEnv("XDG_CONFIG_HOME", configDir)
	defer cleanup1()

	cacheDir := os.TempDir()
	cleanup2 := setEnv("XDG_CACHE_HOME", cacheDir)
	defer cleanup2()

	goleak.VerifyTestMain(m)
}

func startServer(t *testing.T, useTLS, useReflection, useWeb bool) (func(), string) {
	t.Helper()

	port, err := freeport.GetFreePort()
	if err != nil {
		t.Fatalf("failed to get a free port for gRPC test server: %s", err)
	}

	addr := fmt.Sprintf(":%d", port)
	opts := []server.Option{server.WithAddr(addr)}
	if useReflection {
		opts = append(opts, server.WithReflection())
	}
	if useTLS {
		opts = append(opts, server.WithTLS())
	}
	if useWeb {
		opts = append(opts, server.WithProtocol(server.ProtocolImprobableGRPCWeb))
	}

	srv := server.New(opts...)
	go srv.Serve()

	return func() {
		srv.Stop()
	}, strconv.Itoa(port)
}

func flatten(s string) string {
	s = strings.Replace(s, "\n", " ", -1)
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(" +")
	return re.ReplaceAllString(s, " ")
}
