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

	"github.com/google/go-cmp/cmp"
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
		panic(fmt.Sprintf("failed to execute 'git rev-parse --show-cdup': %s", err))
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

	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("github.com/desertbit/timer.timerRoutine"))
}

func startServer(t *testing.T, useTLS, useReflection, useWeb, registerEmptyPackageService bool) (func(), string) {
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
	if registerEmptyPackageService {
		opts = append(opts, server.WithEmptyPackageService())
	}

	srv := server.New(opts...)
	go srv.Serve()

	return func() {
		if err := srv.Stop(); err != nil {
			t.Fatalf("Stop must not return an error, but got '%s'", err)
		}
	}, strconv.Itoa(port)
}

func flatten(s string) string {
	s = strings.Replace(s, "\n", " ", -1)
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(" +")
	return re.ReplaceAllString(s, " ")
}

func compareWithGolden(t *testing.T, actual string) {
	t.Helper()

	name := t.Name()
	normalizeFilename := func(name string) string {
		fname := goldenPathReplacer.Replace(strings.ToLower(name)) + ".golden"
		return filepath.Join("testdata", "fixtures", fname)
	}

	fname := normalizeFilename(name)

	if *update {
		if err := ioutil.WriteFile(fname, []byte(actual), 0600); err != nil {
			t.Fatalf("failed to update the golden file: %s", err)
		}
		return
	}

	// Load the golden file.
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Fatalf("failed to load a golden file: %s", err)
	}
	expected := goldenReplacer.Replace(string(b))

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("wrong result: \n%s", diff)
	}
}
