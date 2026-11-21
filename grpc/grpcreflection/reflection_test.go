package grpcreflection

import (
	"context"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestNewClientReflectionVersions(t *testing.T) {
	cases := map[string]struct {
		versions []string
		want     string
	}{
		"v1 only":      {versions: []string{"v1"}, want: "v1"},
		"alpha only":   {versions: []string{"v1alpha"}, want: "v1alpha"},
		"v1 preferred": {versions: []string{"v1", "v1alpha"}, want: "v1"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			methods := make(chan string, 2)
			headers := map[string][]string{"X-Reflection-Test": {"first", "second"}}
			server := grpc.NewServer(grpc.StreamInterceptor(func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				md, _ := metadata.FromIncomingContext(stream.Context())
				if diff := cmp.Diff(headers["X-Reflection-Test"], md.Get("x-reflection-test")); diff != "" {
					return status.Errorf(codes.Unauthenticated, "unexpected reflection headers (-want +got): %s", diff)
				}
				methods <- info.FullMethod
				return handler(srv, stream)
			}))
			t.Cleanup(server.Stop)
			grpc_health_v1.RegisterHealthServer(server, health.NewServer())

			wantServices := []string{"grpc.health.v1.Health"}
			for _, version := range c.versions {
				// The pinned gRPC release only exposes an alpha server. The two
				// reflection protocols are wire-compatible, so serve the same
				// handler at each version's fully qualified service name.
				desc := grpc_reflection_v1alpha.ServerReflection_ServiceDesc
				desc.ServiceName = "grpc.reflection." + version + ".ServerReflection"
				server.RegisterService(&desc, reflection.NewServer(reflection.ServerOptions{Services: server}))
				wantServices = append(wantServices, desc.ServiceName)
			}
			sort.Strings(wantServices)

			listener := bufconn.Listen(1024 * 1024)
			t.Cleanup(func() { listener.Close() })
			go server.Serve(listener)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithBlock(),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
					return listener.Dial()
				}))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { conn.Close() })
			reflectionClient := NewClient(conn, headers)
			t.Cleanup(reflectionClient.Reset)
			// Resolve from the server rather than the process-wide descriptor cache.
			reflectionClient.(*client).resolver = new(protoregistry.Files)

			for i := 0; i < 2; i++ {
				services, err := reflectionClient.ListServices()
				if err != nil {
					t.Fatal(err)
				}
				sort.Strings(services)
				if diff := cmp.Diff(wantServices, services); diff != "" {
					t.Fatalf("unexpected services (-want +got): %s", diff)
				}
				wantMethod := "/grpc.reflection." + c.want + ".ServerReflection/ServerReflectionInfo"
				if got := <-methods; got != wantMethod {
					t.Fatalf("expected reflection method %q, got %q", wantMethod, got)
				}
				if i == 0 {
					desc, err := reflectionClient.FindSymbol("grpc.health.v1.Health")
					if err != nil {
						t.Fatal(err)
					}
					if desc.FullName() != "grpc.health.v1.Health" {
						t.Fatalf("unexpected descriptor: %s", desc.FullName())
					}
				}
				reflectionClient.Reset()
			}
		})
	}
}
