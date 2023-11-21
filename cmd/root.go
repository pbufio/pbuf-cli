package cmd

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/jdx/go-netrc"
	v1 "github.com/pbufio/pbuf-cli/gen/pbuf-registry/v1"
	"github.com/pbufio/pbuf-cli/internal/model"
	"github.com/pbufio/pbuf-cli/internal/modules"
	"github.com/pbufio/pbuf-cli/internal/registry"
	"github.com/spf13/cobra"
)

// NewRootCmd creates cobra root command via function
func NewRootCmd() *cobra.Command {
	// create root command
	rootCmd := &cobra.Command{
		Use:   "pbuf-cli",
		Short: "PowerBuf CLI",
		Long:  "PowerBuf CLI is a command line interface for PowerBuf",
		Run: func(cmd *cobra.Command, args []string) {
			// do nothing
		},
	}

	const modulesConfigFilename = "pbuf.yaml"
	// read the file (modulesConfigFilename) and call ModulesConfig.Vendor()
	file, err := os.ReadFile(modulesConfigFilename)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	modulesConfig, err := modules.NewConfig(file)
	if err != nil {
		log.Fatalf("failed to create modules config: %v", err)
	}

	var netrcAuth *netrc.Netrc

	usr, err := user.Current()
	if err == nil {
		netrcAuth, err = netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
		if err != nil {
			log.Printf("no .netrc file found. skipping auth")
		}
	} else {
		log.Printf("failed to fetch current user: %v", err)
	}

	if modulesConfig.HasRegistry() {
		var registryClient v1.RegistryClient
		if modulesConfig.Registry.Insecure {
			registryClient = registry.NewInsecureClient(modulesConfig, netrcAuth)
		} else {
			registryClient = registry.NewSecureClient(modulesConfig, netrcAuth)
		}

		rootCmd.AddCommand(NewModuleCmd(modulesConfig, registryClient))
		rootCmd.AddCommand(NewVendorCmd(modulesConfig, netrcAuth, registryClient))
	} else {
		rootCmd.AddCommand(NewVendorCmd(modulesConfig, netrcAuth, nil))
	}

	return rootCmd
}

func NewModuleCmd(config *model.Config, client v1.RegistryClient) *cobra.Command {
	// create module command
	moduleCmd := &cobra.Command{
		Use:   "modules",
		Short: "Modules",
		Long:  "Modules is a command to manage modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// add subcommands
	moduleCmd.AddCommand(NewRegisterModuleCmd(config, client))
	moduleCmd.AddCommand(NewPushModuleCmd(config, client))
	moduleCmd.AddCommand(NewDeleteTagCmd(config, client))
	moduleCmd.AddCommand(NewDeleteModuleCmd(config, client))

	moduleCmd.AddCommand(NewListModulesCmd(config, client))
	moduleCmd.AddCommand(NewGetModuleCmd(config, client))

	return moduleCmd
}

func NewRegisterModuleCmd(config *model.Config, client v1.RegistryClient) *cobra.Command {
	// create register module command
	registerModuleCmd := &cobra.Command{
		Use:   "register",
		Short: "Register",
		Long:  "Register is a command to register modules",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if config.Name == "" {
				log.Fatalf("module name is required. see pbuf.yaml reference")
			}

			module, err := client.RegisterModule(cmd.Context(), &v1.RegisterModuleRequest{
				Name: config.Name,
			})

			if err != nil {
				log.Fatalf("failed to register: %v", err)
			}

			log.Printf("module %s successfully registered", module.Name)
		},
	}

	return registerModuleCmd
}

func NewPushModuleCmd(config *model.Config, client v1.RegistryClient) *cobra.Command {
	// create push module command
	pushModuleCmd := &cobra.Command{
		Use:   "push [tag]",
		Short: "Push",
		Long:  "Push is a command to push modules",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if config.Name == "" {
				log.Fatalf("module name is required. see pbuf.yaml reference")
			}

			tag := args[0]

			protoFiles, err := registry.CollectProtoFilesInDirs(config.Export.Paths)
			if err != nil {
				log.Fatalf("failed to collect proto files: %v", err)
			}

			var dependencies []*v1.Dependency
			for _, dependency := range config.Modules {
				if dependency.Name != "" && dependency.Repository == "" {
					dependencies = append(dependencies, &v1.Dependency{
						Name: dependency.Name,
						Tag:  dependency.Tag,
					})
				}
			}

			log.Printf("pushing module %s with tag %s", config.Name, tag)

			module, err := client.PushModule(cmd.Context(), &v1.PushModuleRequest{
				ModuleName:   config.Name,
				Tag:          tag,
				Protofiles:   protoFiles,
				Dependencies: dependencies,
			})

			if err != nil {
				log.Fatalf("failed to push: %v", err)
			}

			log.Printf("module %s successfully pushed", module.Name)
		},
	}

	return pushModuleCmd
}

func NewDeleteTagCmd(config *model.Config, client v1.RegistryClient) *cobra.Command {
	// create delete tag command
	deleteTagCmd := &cobra.Command{
		Use:   "delete-tag [tag]",
		Short: "Delete tag",
		Long:  "Delete tag is a command to delete tags",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if config.Name == "" {
				log.Fatalf("module name is required. see pbuf.yaml reference")
			}

			tag := args[0]

			if tag == "" {
				log.Fatalf("tag is required")
			}

			_, err := client.DeleteModuleTag(cmd.Context(), &v1.DeleteModuleTagRequest{
				Name: config.Name,
				Tag:  tag,
			})

			if err != nil {
				log.Fatalf("failed to delete tag: %v", err)
			}

			log.Printf("tag %s successfully deleted", tag)
		},
	}

	return deleteTagCmd
}

func NewDeleteModuleCmd(config *model.Config, client v1.RegistryClient) *cobra.Command {
	// create delete module command
	deleteModuleCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete",
		Long:  "Delete is a command to delete modules",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if config.Name == "" {
				log.Fatalf("module name is required. see pbuf.yaml reference")
			}

			_, err := client.DeleteModule(cmd.Context(), &v1.DeleteModuleRequest{
				Name: config.Name,
			})

			if err != nil {
				log.Fatalf("failed to delete: %v", err)
			}

			log.Printf("module %s successfully deleted", config.Name)
		},
	}

	return deleteModuleCmd
}

// NewGetModuleCmd creates cobra command for get module
func NewGetModuleCmd(config *model.Config, client v1.RegistryClient) *cobra.Command {
	// create get module command
	getModuleCmd := &cobra.Command{
		Use:   "get [module_name]",
		Short: "Get",
		Long:  "Get is a command to get modules",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moduleName := args[0]

			module, err := client.GetModule(cmd.Context(), &v1.GetModuleRequest{
				Name: moduleName,
			})

			if err != nil {
				log.Fatalf("failed to get: %v", err)
			}

			moduleDependencies, err := client.GetModuleDependencies(cmd.Context(), &v1.GetModuleDependenciesRequest{
				Name: moduleName,
			})

			if err != nil {
				log.Fatalf("failed to fetch dependencies: %v", err)
			}

			// print as json pretty
			marshalled, err := json.MarshalIndent(module, "", "  ")
			if err != nil {
				log.Fatalf("failed to marshal module: %v", err)
			}
			log.Printf("module:\n%+v\n", string(marshalled))

			// print deps as json pretty
			marshalled, err = json.MarshalIndent(moduleDependencies, "", "  ")
			if err != nil {
				log.Fatalf("failed to marshal module dependencies: %v", err)
			}
			log.Printf("module dependencies:\n%+v\n", string(marshalled))
		},
	}

	return getModuleCmd
}

// NewListModulesCmd creates cobra command for list modules
func NewListModulesCmd(config *model.Config, client v1.RegistryClient) *cobra.Command {
	// create list modules command
	listModulesCmd := &cobra.Command{
		Use:   "list",
		Short: "List",
		Long:  "List is a command to list modules",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			modulesList, err := client.ListModules(cmd.Context(), &v1.ListModulesRequest{})
			if err != nil {
				log.Fatalf("failed to fetch modules list: %v", err)
			}

			// print as json pretty
			marshalled, err := json.MarshalIndent(modulesList, "", "  ")
			if err != nil {
				log.Fatalf("failed to marshal modules list: %v", err)
			}
			log.Printf("%+v", string(marshalled))
		},
	}

	return listModulesCmd
}

// NewVendorCmd creates cobra command for vendor
func NewVendorCmd(modulesConfig *model.Config, netrcAuth *netrc.Netrc, client v1.RegistryClient) *cobra.Command {
	// create vendor command
	vendorCmd := &cobra.Command{
		Use:   "vendor",
		Short: "Vendor",
		Long:  "Vendor is a command to vendor modules",
		Run: func(cmd *cobra.Command, args []string) {
			err := modules.Vendor(modulesConfig, netrcAuth, client)
			if err != nil {
				log.Fatalf("failed to vendor: %v", err)
			}
		},
	}

	return vendorCmd
}
