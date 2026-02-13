package cmd

import (
	v1 "github.com/pbufio/pbuf-cli/gen/pbuf-registry/v1"
	"github.com/pbufio/pbuf-cli/internal/model"
	"github.com/spf13/cobra"
)

func NewDriftCmd(_ *model.Config, client v1.DriftServiceClient) *cobra.Command {
	driftCmd := &cobra.Command{
		Use:   "drift",
		Short: "Drift",
		Long:  "Drift is a command to manage drift detection events",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	driftCmd.AddCommand(newListDriftEventsCmd(client))
	driftCmd.AddCommand(newGetModuleDriftEventsCmd(client))

	return driftCmd
}

func newListDriftEventsCmd(client v1.DriftServiceClient) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List drift events",
		Long:  "List is a command to list all drift events",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			unacknowledgedOnly, err := cmd.Flags().GetBool("unacknowledged-only")
			if err != nil {
				return err
			}

			resp, err := client.ListDriftEvents(cmd.Context(), &v1.ListDriftEventsRequest{
				UnacknowledgedOnly: unacknowledgedOnly,
			})
			if err != nil {
				return err
			}

			return printJSON(resp)
		},
	}

	listCmd.Flags().Bool("unacknowledged-only", true, "only return unacknowledged events")
	return listCmd
}

func newGetModuleDriftEventsCmd(client v1.DriftServiceClient) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "module [module_name]",
		Short: "Get module drift events",
		Long:  "Module is a command to get drift events for a specific module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			tagName, err := cmd.Flags().GetString("tag")
			if err != nil {
				return err
			}

			req := &v1.GetModuleDriftEventsRequest{
				ModuleName: moduleName,
			}
			if tagName != "" {
				req.TagName = &tagName
			}

			resp, err := client.GetModuleDriftEvents(cmd.Context(), req)
			if err != nil {
				return err
			}

			return printJSON(resp)
		},
	}

	getCmd.Flags().String("tag", "", "filter drift events by tag name")
	return getCmd
}
