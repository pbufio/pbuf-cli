package e2e

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/pbufio/pbuf-cli/gen/pbuf-registry/v1"
	"github.com/pbufio/pbuf-cli/internal/model"
	"github.com/pbufio/pbuf-cli/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// testRegistryServer implements the Registry gRPC server for integration testing
type testRegistryServer struct {
	v1.UnimplementedRegistryServer
	modules map[string]*moduleData
}

type moduleData struct {
	module       *v1.Module
	protoFiles   []*v1.ProtoFile
	dependencies []*v1.Dependency
}

func newTestRegistryServer() *testRegistryServer {
	return &testRegistryServer{
		modules: make(map[string]*moduleData),
	}
}

func (s *testRegistryServer) addModule(name, tag string, files []*v1.ProtoFile, deps []*v1.Dependency) {
	key := name + "@" + tag
	s.modules[key] = &moduleData{
		module:       &v1.Module{Name: name, Tags: []string{tag}},
		protoFiles:   files,
		dependencies: deps,
	}
}

func (s *testRegistryServer) PullModule(_ context.Context, req *v1.PullModuleRequest) (*v1.PullModuleResponse, error) {
	key := req.Name + "@" + req.Tag
	data, ok := s.modules[key]
	if !ok {
		return nil, fmt.Errorf("module %s not found", key)
	}
	return &v1.PullModuleResponse{
		Module:     data.module,
		Protofiles: data.protoFiles,
	}, nil
}

func (s *testRegistryServer) GetModuleDependencies(_ context.Context, req *v1.GetModuleDependenciesRequest) (*v1.GetModuleDependenciesResponse, error) {
	key := req.Name + "@" + req.Tag
	data, ok := s.modules[key]
	if !ok {
		return &v1.GetModuleDependenciesResponse{}, nil
	}

	if req.ResolveTransitive {
		// Return all dependencies including transitive
		return &v1.GetModuleDependenciesResponse{
			Dependencies: data.dependencies,
		}, nil
	}

	// Return only direct dependencies
	var direct []*v1.Dependency
	for _, dep := range data.dependencies {
		if dep.DependencyType == "direct" {
			direct = append(direct, dep)
		}
	}
	return &v1.GetModuleDependenciesResponse{
		Dependencies: direct,
	}, nil
}

func (s *testRegistryServer) ListModules(_ context.Context, _ *v1.ListModulesRequest) (*v1.ListModulesResponse, error) {
	return &v1.ListModulesResponse{}, nil
}

func (s *testRegistryServer) GetModule(_ context.Context, _ *v1.GetModuleRequest) (*v1.Module, error) {
	return &v1.Module{}, nil
}

func (s *testRegistryServer) RegisterModule(_ context.Context, _ *v1.RegisterModuleRequest) (*v1.Module, error) {
	return &v1.Module{}, nil
}

func (s *testRegistryServer) PushModule(_ context.Context, _ *v1.PushModuleRequest) (*v1.Module, error) {
	return &v1.Module{}, nil
}

func (s *testRegistryServer) DeleteModule(_ context.Context, _ *v1.DeleteModuleRequest) (*v1.DeleteModuleResponse, error) {
	return &v1.DeleteModuleResponse{}, nil
}

func (s *testRegistryServer) DeleteModuleTag(_ context.Context, _ *v1.DeleteModuleTagRequest) (*v1.DeleteModuleTagResponse, error) {
	return &v1.DeleteModuleTagResponse{}, nil
}

// startTestServer starts a gRPC server on a random port and returns the connection and cleanup function
func startTestServer(t *testing.T, server *testRegistryServer) (v1.RegistryClient, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	v1.RegisterRegistryServer(grpcServer, server)

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := v1.NewRegistryClient(conn)
	cleanup := func() {
		_ = conn.Close()
		grpcServer.Stop()
	}

	return client, cleanup
}

func TestIntegration_VendorWithTransitiveDependencies(t *testing.T) {
	// Setup test server with module hierarchy:
	// module-a depends on module-b (direct)
	// module-b depends on module-c (transitive from A's perspective)
	server := newTestRegistryServer()

	server.addModule("org/module-a", "v1.0.0",
		[]*v1.ProtoFile{
			{Filename: "api/a/service.proto", Content: `syntax = "proto3";
package a;
import "api/b/common.proto";
message ARequest { b.Common common = 1; }
`},
		},
		[]*v1.Dependency{
			{Name: "org/module-b", Tag: "v1.0.0", DependencyType: "direct"},
			{Name: "org/module-c", Tag: "v2.0.0", DependencyType: "transitive"},
		},
	)

	server.addModule("org/module-b", "v1.0.0",
		[]*v1.ProtoFile{
			{Filename: "api/b/common.proto", Content: `syntax = "proto3";
package b;
message Common { string value = 1; }
`},
		},
		[]*v1.Dependency{
			{Name: "org/module-c", Tag: "v2.0.0", DependencyType: "direct"},
		},
	)

	server.addModule("org/module-c", "v2.0.0",
		[]*v1.ProtoFile{
			{Filename: "api/c/types.proto", Content: `syntax = "proto3";
package c;
message Type { string name = 1; }
`},
		},
		nil,
	)

	client, cleanup := startTestServer(t, server)
	defer cleanup()

	// Create temp directory and change to it
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }()

	// Configure: user only lists module-a as dependency
	config := &model.Config{
		Version:  "v1",
		Name:     "org/my-project",
		Registry: model.Registry{Addr: "127.0.0.1:6777"},
		Modules: []*model.Module{
			{
				Name:         "org/module-a",
				Tag:          "v1.0.0",
				Path:         "api/a",
				OutputFolder: "vendor/a",
			},
		},
	}

	// Vendor modules
	err = modules.Vendor(config, nil, client)
	require.NoError(t, err)

	// Verify module-a was vendored
	aProto := filepath.Join(tmpDir, "vendor", "a", "service.proto")
	content, err := os.ReadFile(aProto)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package a;")
	assert.Contains(t, string(content), "ARequest")

	// Verify module-b was vendored as a direct transitive dependency
	bProto := filepath.Join(tmpDir, "api", "b", "common.proto")
	content, err = os.ReadFile(bProto)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package b;")
	assert.Contains(t, string(content), "Common")

	// Verify module-c was NOT vendored (it's a transitive dependency, not direct)
	cProto := filepath.Join(tmpDir, "api", "c", "types.proto")
	_, err = os.Stat(cProto)
	assert.True(t, os.IsNotExist(err), "transitive dependency module-c should not be vendored")
}

func TestIntegration_VendorSkipsDuplicates(t *testing.T) {
	// Test that if module-b is already in config AND is a transitive dep of module-a,
	// it's only vendored once (from config, not again as transitive)
	server := newTestRegistryServer()

	server.addModule("org/module-a", "v1.0.0",
		[]*v1.ProtoFile{
			{Filename: "a.proto", Content: `syntax = "proto3"; package a;`},
		},
		[]*v1.Dependency{
			{Name: "org/module-b", Tag: "v1.0.0", DependencyType: "direct"},
		},
	)

	server.addModule("org/module-b", "v1.0.0",
		[]*v1.ProtoFile{
			{Filename: "b.proto", Content: `syntax = "proto3"; package b;`},
		},
		nil,
	)

	// Wrap server to count pulls
	countingServer := &countingRegistryServer{
		testRegistryServer: server,
		pullCounts:         make(map[string]int),
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	v1.RegisterRegistryServer(grpcServer, countingServer)

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	client := v1.NewRegistryClient(conn)

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }()

	// Both module-a and module-b in config
	config := &model.Config{
		Version:  "v1",
		Name:     "org/my-project",
		Registry: model.Registry{Addr: "127.0.0.1:6777"},
		Modules: []*model.Module{
			{Name: "org/module-a", Tag: "v1.0.0"},
			{Name: "org/module-b", Tag: "v1.0.0"},
		},
	}

	err = modules.Vendor(config, nil, client)
	require.NoError(t, err)

	// module-b should only be pulled once (from config), not again as transitive
	assert.Equal(t, 1, countingServer.pullCounts["org/module-b@v1.0.0"])
}

type countingRegistryServer struct {
	*testRegistryServer
	pullCounts map[string]int
}

func (s *countingRegistryServer) PullModule(ctx context.Context, req *v1.PullModuleRequest) (*v1.PullModuleResponse, error) {
	key := req.Name + "@" + req.Tag
	s.pullCounts[key]++
	return s.testRegistryServer.PullModule(ctx, req)
}

func TestIntegration_PushWithDependencyType(t *testing.T) {
	// Test that push sends dependency_type="direct" for configured modules
	server := newTestRegistryServer()

	var capturedReq *v1.PushModuleRequest

	// Create a custom server that captures the push request
	capturingServer := &capturingRegistryServer{
		testRegistryServer: server,
		onPush: func(req *v1.PushModuleRequest) {
			capturedReq = req
		},
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	v1.RegisterRegistryServer(grpcServer, capturingServer)

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	client := v1.NewRegistryClient(conn)

	// Push module with dependencies
	_, err = client.PushModule(context.Background(), &v1.PushModuleRequest{
		ModuleName: "org/my-module",
		Tag:        "v1.0.0",
		Dependencies: []*v1.Dependency{
			{Name: "org/dep-a", Tag: "v1.0.0", DependencyType: "direct"},
			{Name: "org/dep-b", Tag: "v2.0.0", DependencyType: "direct"},
		},
	})
	require.NoError(t, err)

	// Verify the push request included dependency types
	require.NotNil(t, capturedReq)
	require.Len(t, capturedReq.Dependencies, 2)
	assert.Equal(t, "direct", capturedReq.Dependencies[0].DependencyType)
	assert.Equal(t, "direct", capturedReq.Dependencies[1].DependencyType)
}

type capturingRegistryServer struct {
	*testRegistryServer
	onPush func(req *v1.PushModuleRequest)
}

func (s *capturingRegistryServer) PushModule(_ context.Context, req *v1.PushModuleRequest) (*v1.Module, error) {
	if s.onPush != nil {
		s.onPush(req)
	}
	return &v1.Module{Name: req.ModuleName, Tags: []string{req.Tag}}, nil
}

func TestIntegration_GetDependenciesWithResolveTransitive(t *testing.T) {
	server := newTestRegistryServer()

	server.addModule("org/module-a", "v1.0.0",
		nil,
		[]*v1.Dependency{
			{Name: "org/module-b", Tag: "v1.0.0", DependencyType: "direct"},
			{Name: "org/module-c", Tag: "v2.0.0", DependencyType: "transitive"},
		},
	)

	client, cleanup := startTestServer(t, server)
	defer cleanup()

	// Without resolve_transitive: only direct deps
	resp, err := client.GetModuleDependencies(context.Background(), &v1.GetModuleDependenciesRequest{
		Name: "org/module-a",
		Tag:  "v1.0.0",
	})
	require.NoError(t, err)
	require.Len(t, resp.Dependencies, 1)
	assert.Equal(t, "org/module-b", resp.Dependencies[0].Name)
	assert.Equal(t, "direct", resp.Dependencies[0].DependencyType)

	// With resolve_transitive: all deps
	resp, err = client.GetModuleDependencies(context.Background(), &v1.GetModuleDependenciesRequest{
		Name:              "org/module-a",
		Tag:               "v1.0.0",
		ResolveTransitive: true,
	})
	require.NoError(t, err)
	require.Len(t, resp.Dependencies, 2)
	assert.Equal(t, "org/module-b", resp.Dependencies[0].Name)
	assert.Equal(t, "direct", resp.Dependencies[0].DependencyType)
	assert.Equal(t, "org/module-c", resp.Dependencies[1].Name)
	assert.Equal(t, "transitive", resp.Dependencies[1].DependencyType)
}
