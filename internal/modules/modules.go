package modules

import (
	"log"
	"os"

	"github.com/jdx/go-netrc"
	v1 "github.com/pbufio/pbuf-cli/gen/pbuf-registry/v1"
	"github.com/pbufio/pbuf-cli/internal/git"
	"github.com/pbufio/pbuf-cli/internal/model"
	"github.com/pbufio/pbuf-cli/internal/patcher"
	"github.com/pbufio/pbuf-cli/internal/registry"
	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v2"
)

// NewConfig create a struct for bytes array
func NewConfig(contents []byte) (*model.Config, error) {
	modulesConfig := &model.Config{}
	err := yaml.Unmarshal(contents, modulesConfig)
	if err != nil {
		return nil, err
	}
	return modulesConfig, nil
}

// newProtoPatchers create a slice of proto patchers
func newProtoPatchers() []patcher.Patcher {
	var result []patcher.Patcher

	// if we have go.mod file
	// then parse it and fetch the module name
	// and pass it to the go package patcher
	file, err := os.ReadFile("go.mod")
	if err == nil {
		// that's ok, we cannot find go mod file
		path := modfile.ModulePath(file)
		if path != "" {
			result = append(result, patcher.NewGoPackagePatcher(path))
		}
	}
	return result
}

// Vendor function that iterate over the modules and vendor proto files from git repositories
func Vendor(config *model.Config, netrcAuth *netrc.Netrc, client v1.RegistryClient) error {
	patchers := newProtoPatchers()

	for _, module := range config.Modules {
		if module.Repository == "" {
			if config.HasRegistry() {
				if module.Name == "" {
					log.Fatalf("no module name found for module: %v", module)
				}

				if module.Tag == "" {
					log.Fatalf("no module tag found for module: %v", module)
				}

				err := registry.VendorRegistryModule(module, client, patchers)
				if err != nil {
					log.Fatalf("failed to vendor module %s: %v", module.Name, err)
				}
			} else {
				log.Fatalf("no repository found for module: %s", module.Name)
			}
		} else {
			err := git.VendorGitModule(module, netrcAuth, patchers)
			if err != nil {
				log.Fatalf("failed to vendor module %s: %v", module.Repository, err)
			}
		}
	}

	return nil
}
