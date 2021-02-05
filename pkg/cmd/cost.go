package cmd

import (
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	// "k8s.io/client-go/tools/clientcmd"
	// "k8s.io/client-go/tools/clientcmd/api"

	"github.com/spf13/cobra"

	"github.com/kubecost/cost-model/pkg/kubecost"
)

// Note that the auth/gcp import is necessary https://github.com/kubernetes/client-go/issues/242
// Similar may be required for AWS

var (
	costExample = `
    # Show the projected monthly rate for each namespace based on the last 5 days of activity.
    %[1]s cost namespace --window 5d

    # Show how much each namespace cost over the past 5 days with additional CPU and memory cost and efficiency breakdown.
    %[1]s cost namespace --historical --window 5d --show-cpu --show-memory --show-efficiency

    # Show the projected monthly rate for each deployment based on the last month of activity with CPU, memory, GPU, PV, and network cost breakdown.
    %[1]s cost deployment --window month --show-cpu --show-memory --show-gpu --show-pv --show-network
`

	errNoContext = fmt.Errorf("no context is currently set, use %q to select a new one", "kubectl config use-context <context>")
)

const (
	idleString = "__idle__"
)

// CommonCostOptions provides information required to get
// cost information from the kubecost API
type CostOptionsCommon struct {
	configFlags *genericclioptions.ConfigFlags

	costWindow string

	restConfig *rest.Config
	args       []string

	genericclioptions.IOStreams
}

// NewCommonCostOptions creates the default set of cost options
func NewCommonCostOptions(streams genericclioptions.IOStreams) *CostOptionsCommon {
	return &CostOptionsCommon{
		configFlags: genericclioptions.NewConfigFlags(true),

		IOStreams: streams,
	}
}

// NewCmdCost provides a cobra command wrapping CostOptions
func NewCmdCost(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewCommonCostOptions(streams)

	cmd := &cobra.Command{
		Use:          "cost",
		Short:        "View cluster cost information.",
		Example:      fmt.Sprintf(costExample, "kubectl"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			// if err := o.Complete(c, args); err != nil {
			// 	return err
			// }
			// if err := o.Validate(); err != nil {
			// 	return err
			// }
			// if err := o.Run(); err != nil {
			// 	return err
			// }
			return fmt.Errorf("please use a subcommand")

			// return nil
		},
	}

	o.configFlags.AddFlags(cmd.Flags())

	cmd.AddCommand(newCmdCostNamespace(streams))
	cmd.AddCommand(newCmdCostDeployment(streams))

	return cmd
}

// Complete sets all information required for getting cost information
func (o *CostOptionsCommon) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	var err error

	o.restConfig, err = o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *CostOptionsCommon) Validate() error {
	if len(o.args) > 1 {
		return fmt.Errorf("either one or no arguments are allowed")
	}

	// just make sure window parses client-side, perhaps not necessary
	if _, err := kubecost.ParseWindowWithOffset(o.costWindow, 0); err != nil {
		return fmt.Errorf("failed to parse window: %s", err)
	}

	return nil
}

func (o *CostOptionsCommon) Run() error {

	clientset, err := kubernetes.NewForConfig(o.restConfig)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %s", err)
	}

	allocResp, err := queryAllocation(clientset, o.costWindow, "")
	if err != nil {
		return fmt.Errorf("failed to query allocation API: %s", err)
	}

	// using allocResp.Data[0] is fine because we set the accumulate
	// flag in the allocation API
	// err = filterAllocations(allocResp.Data[0], o.costNamespace)
	// if err != nil {
	// 	return fmt.Errorf("failed to filter allocations: %s", err)
	// }
	writeAllocationTable(o.Out, allocResp.Data[0])

	return nil
}