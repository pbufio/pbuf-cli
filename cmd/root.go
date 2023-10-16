package cmd

import (
	"log"
	"os"

	"github.com/pbufio/pbuf-cli/internal/modules"
	"github.com/spf13/cobra"
)

// NewRootCmd creates cobra root command via function
func NewRootCmd() *cobra.Command {
	// create root command
	rootCmd := &cobra.Command{
		Use:   "pbuf",
		Short: "PowerBuf CLI",
		Long:  "PowerBuf CLI is a command line interface for PowerBuf",
		Run: func(cmd *cobra.Command, args []string) {
			// do nothing
		},
	}
	// add subcommands
	rootCmd.AddCommand(NewVendorCmd())

	return rootCmd
}

// NewVendorCmd creates cobra command for vendor
func NewVendorCmd() *cobra.Command {
	const modulesConfigFilename = "pbuf.yaml"
	// create vendor command
	vendorCmd := &cobra.Command{
		Use:   "vendor",
		Short: "Vendor",
		Long:  "Vendor is a command to vendor modules",
		Run: func(cmd *cobra.Command, args []string) {
			// read the file (modulesConfigFilename) and call ModulesConfig.Vendor()
			file, err := os.ReadFile(modulesConfigFilename)
			if err != nil {
				log.Fatalf("failed to read file: %v", err)
			}

			modulesConfig, err := modules.NewConfig(file)
			if err != nil {
				log.Fatalf("failed to create modules config: %v", err)
			}

			err = modules.Vendor(modulesConfig)
			if err != nil {
				log.Fatalf("failed to vendor: %v", err)
			}
		},
	}

	return vendorCmd
}
