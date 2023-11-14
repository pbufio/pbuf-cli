package modules

import (
	"log"
	"os/user"
	"path/filepath"

	"github.com/jdx/go-netrc"
	v1 "github.com/pbufio/pbuf-cli/gen/pbuf-registry/v1"
	"github.com/pbufio/pbuf-cli/internal/git"
	"github.com/pbufio/pbuf-cli/internal/model"
	"github.com/pbufio/pbuf-cli/internal/registry"
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

// Vendor function that iterate over the modules and vendor proto files from git repositories
func Vendor(config *model.Config, client v1.RegistryClient) error {
	usr, err := user.Current()
	if err != nil {
		log.Printf("failed to get current user")
		return err
	}

	netrcAuth, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	if err != nil {
		log.Printf("no .netrc file found. skipping auth")
	}

	for _, module := range config.Modules {
		if module.Repository == "" {
			if config.HasRegistry() {
				if module.Name == "" {
					log.Fatalf("no module name found for module: %v", module)
				}

				if module.Tag == "" {
					log.Fatalf("no module tag found for module: %v", module)
				}

				err := registry.VendorRegistryModule(module, client)
				if err != nil {
					log.Fatalf("failed to vendor module %s: %v", module.Name, err)
				}
			} else {
				log.Fatalf("no repository found for module: %s", module.Name)
			}
		} else {
			err := git.VendorGitModule(module, netrcAuth)
			if err != nil {
				log.Fatalf("failed to vendor module %s: %v", module.Repository, err)
			}
		}
	}

	return nil
}
