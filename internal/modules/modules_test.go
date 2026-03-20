package modules

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/pbufio/pbuf-cli/gen/pbuf-registry/v1"
	"github.com/pbufio/pbuf-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// mockRegistryClient implements v1.RegistryClient for testing
type mockRegistryClient struct {
	pullModuleFunc           func(ctx context.Context, in *v1.PullModuleRequest, opts ...grpc.CallOption) (*v1.PullModuleResponse, error)
	getModuleDependenciesFunc func(ctx context.Context, in *v1.GetModuleDependenciesRequest, opts ...grpc.CallOption) (*v1.GetModuleDependenciesResponse, error)
}

func (m *mockRegistryClient) ListModules(ctx context.Context, in *v1.ListModulesRequest, opts ...grpc.CallOption) (*v1.ListModulesResponse, error) {
	return nil, nil
}

func (m *mockRegistryClient) GetModule(ctx context.Context, in *v1.GetModuleRequest, opts ...grpc.CallOption) (*v1.Module, error) {
	return nil, nil
}

func (m *mockRegistryClient) RegisterModule(ctx context.Context, in *v1.RegisterModuleRequest, opts ...grpc.CallOption) (*v1.Module, error) {
	return nil, nil
}

func (m *mockRegistryClient) PullModule(ctx context.Context, in *v1.PullModuleRequest, opts ...grpc.CallOption) (*v1.PullModuleResponse, error) {
	if m.pullModuleFunc != nil {
		return m.pullModuleFunc(ctx, in, opts...)
	}
	return &v1.PullModuleResponse{}, nil
}

func (m *mockRegistryClient) PushModule(ctx context.Context, in *v1.PushModuleRequest, opts ...grpc.CallOption) (*v1.Module, error) {
	return nil, nil
}

func (m *mockRegistryClient) DeleteModule(ctx context.Context, in *v1.DeleteModuleRequest, opts ...grpc.CallOption) (*v1.DeleteModuleResponse, error) {
	return nil, nil
}

func (m *mockRegistryClient) DeleteModuleTag(ctx context.Context, in *v1.DeleteModuleTagRequest, opts ...grpc.CallOption) (*v1.DeleteModuleTagResponse, error) {
	return nil, nil
}

func (m *mockRegistryClient) GetModuleDependencies(ctx context.Context, in *v1.GetModuleDependenciesRequest, opts ...grpc.CallOption) (*v1.GetModuleDependenciesResponse, error) {
	if m.getModuleDependenciesFunc != nil {
		return m.getModuleDependenciesFunc(ctx, in, opts...)
	}
	return &v1.GetModuleDependenciesResponse{}, nil
}

func TestResolveTransitiveDependencies_NoDeps(t *testing.T) {
	client := &mockRegistryClient{
		getModuleDependenciesFunc: func(ctx context.Context, in *v1.GetModuleDependenciesRequest, opts ...grpc.CallOption) (*v1.GetModuleDependenciesResponse, error) {
			assert.True(t, in.ResolveTransitive)
			return &v1.GetModuleDependenciesResponse{
				Dependencies: []*v1.Dependency{},
			}, nil
		},
	}

	config := &model.Config{
		Registry: model.Registry{Addr: "test.registry:6777"},
		Modules: []*model.Module{
			{Name: "module-a", Tag: "v1.0.0"},
		},
	}

	deps := resolveTransitiveDependencies(config, client)
	assert.Empty(t, deps)
}

func TestResolveTransitiveDependencies_DirectDeps(t *testing.T) {
	client := &mockRegistryClient{
		getModuleDependenciesFunc: func(ctx context.Context, in *v1.GetModuleDependenciesRequest, opts ...grpc.CallOption) (*v1.GetModuleDependenciesResponse, error) {
			assert.True(t, in.ResolveTransitive)
			return &v1.GetModuleDependenciesResponse{
				Dependencies: []*v1.Dependency{
					{Name: "module-b", Tag: "v1.0.0", DependencyType: "direct"},
					{Name: "module-c", Tag: "v2.0.0", DependencyType: "transitive"},
				},
			}, nil
		},
	}

	config := &model.Config{
		Registry: model.Registry{Addr: "test.registry:6777"},
		Modules: []*model.Module{
			{Name: "module-a", Tag: "v1.0.0"},
		},
	}

	deps := resolveTransitiveDependencies(config, client)
	require.Len(t, deps, 1)
	assert.Equal(t, "module-b", deps[0].Name)
	assert.Equal(t, "v1.0.0", deps[0].Tag)
}

func TestResolveTransitiveDependencies_SkipsConfigured(t *testing.T) {
	client := &mockRegistryClient{
		getModuleDependenciesFunc: func(ctx context.Context, in *v1.GetModuleDependenciesRequest, opts ...grpc.CallOption) (*v1.GetModuleDependenciesResponse, error) {
			return &v1.GetModuleDependenciesResponse{
				Dependencies: []*v1.Dependency{
					{Name: "module-b", Tag: "v1.0.0", DependencyType: "direct"},
					{Name: "module-c", Tag: "v2.0.0", DependencyType: "direct"},
				},
			}, nil
		},
	}

	config := &model.Config{
		Registry: model.Registry{Addr: "test.registry:6777"},
		Modules: []*model.Module{
			{Name: "module-a", Tag: "v1.0.0"},
			{Name: "module-b", Tag: "v1.0.0"},
		},
	}

	deps := resolveTransitiveDependencies(config, client)
	require.Len(t, deps, 1)
	assert.Equal(t, "module-c", deps[0].Name)
}

func TestResolveTransitiveDependencies_SkipsGitModules(t *testing.T) {
	called := false
	client := &mockRegistryClient{
		getModuleDependenciesFunc: func(ctx context.Context, in *v1.GetModuleDependenciesRequest, opts ...grpc.CallOption) (*v1.GetModuleDependenciesResponse, error) {
			called = true
			return &v1.GetModuleDependenciesResponse{}, nil
		},
	}

	config := &model.Config{
		Registry: model.Registry{Addr: "test.registry:6777"},
		Modules: []*model.Module{
			{Repository: "https://github.com/example/repo", Branch: "main"},
		},
	}

	deps := resolveTransitiveDependencies(config, client)
	assert.Empty(t, deps)
	assert.False(t, called)
}

func TestResolveTransitiveDependencies_Deduplicates(t *testing.T) {
	client := &mockRegistryClient{
		getModuleDependenciesFunc: func(ctx context.Context, in *v1.GetModuleDependenciesRequest, opts ...grpc.CallOption) (*v1.GetModuleDependenciesResponse, error) {
			return &v1.GetModuleDependenciesResponse{
				Dependencies: []*v1.Dependency{
					{Name: "shared-dep", Tag: "v1.0.0", DependencyType: "direct"},
				},
			}, nil
		},
	}

	config := &model.Config{
		Registry: model.Registry{Addr: "test.registry:6777"},
		Modules: []*model.Module{
			{Name: "module-a", Tag: "v1.0.0"},
			{Name: "module-b", Tag: "v2.0.0"},
		},
	}

	deps := resolveTransitiveDependencies(config, client)
	require.Len(t, deps, 1)
	assert.Equal(t, "shared-dep", deps[0].Name)
}

func TestVendor_WithTransitiveDeps(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }()

	protoContent := `syntax = "proto3";
package test;
message TestMsg {}
`

	client := &mockRegistryClient{
		pullModuleFunc: func(ctx context.Context, in *v1.PullModuleRequest, opts ...grpc.CallOption) (*v1.PullModuleResponse, error) {
			return &v1.PullModuleResponse{
				Module: &v1.Module{Name: in.Name},
				Protofiles: []*v1.ProtoFile{
					{Filename: "test.proto", Content: protoContent},
				},
			}, nil
		},
		getModuleDependenciesFunc: func(ctx context.Context, in *v1.GetModuleDependenciesRequest, opts ...grpc.CallOption) (*v1.GetModuleDependenciesResponse, error) {
			if in.Name == "module-a" {
				return &v1.GetModuleDependenciesResponse{
					Dependencies: []*v1.Dependency{
						{Name: "module-b", Tag: "v1.0.0", DependencyType: "direct"},
					},
				}, nil
			}
			return &v1.GetModuleDependenciesResponse{}, nil
		},
	}

	config := &model.Config{
		Registry: model.Registry{Addr: "test.registry:6777"},
		Modules: []*model.Module{
			{Name: "module-a", Tag: "v1.0.0", OutputFolder: "out-a"},
		},
	}

	err := Vendor(config, nil, client)
	require.NoError(t, err)

	// Check that the direct module was vendored
	_, err = os.Stat(filepath.Join(tmpDir, "out-a", "test.proto"))
	assert.NoError(t, err)

	// Check that the transitive dependency was also vendored
	_, err = os.Stat(filepath.Join(tmpDir, "test.proto"))
	assert.NoError(t, err)
}

func TestVendor_NoTransitiveForGitModules(t *testing.T) {
	// Verify that git-only configs don't trigger transitive resolution
	config := &model.Config{
		Modules: []*model.Module{
			{Repository: "https://github.com/example/repo", Branch: "main"},
		},
	}

	// This should not panic or error - git modules don't have transitive deps
	// We can't actually vendor git modules in test, but verify config parsing works
	assert.False(t, config.HasRegistry())
}

func TestNewConfig(t *testing.T) {
	yamlContent := `
version: v1
name: test-module
registry:
  addr: test.registry:6777
modules:
  - name: dep-a
    tag: v1.0.0
    out: proto
  - repository: https://github.com/example/repo
    branch: main
    out: external
`
	config, err := NewConfig([]byte(yamlContent))
	require.NoError(t, err)
	assert.Equal(t, "v1", config.Version)
	assert.Equal(t, "test-module", config.Name)
	assert.Equal(t, "test.registry:6777", config.Registry.Addr)
	require.Len(t, config.Modules, 2)
	assert.Equal(t, "dep-a", config.Modules[0].Name)
	assert.Equal(t, "v1.0.0", config.Modules[0].Tag)
	assert.Equal(t, "https://github.com/example/repo", config.Modules[1].Repository)
}
