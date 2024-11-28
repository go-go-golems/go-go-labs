package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/spf13/cobra"
)

var (
	showOutputs   bool
	showResources bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tf-show",
		Short: "Show Terraform state information",
		Run:   run,
	}

	rootCmd.Flags().BoolVar(&showOutputs, "outputs", true, "Show terraform outputs")
	rootCmd.Flags().BoolVar(&showResources, "resources", false, "Show terraform resources")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	execPath := "terraform"
	workingDir := "."

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	state, err := tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	fmt.Println("Format Version:", state.FormatVersion)
	fmt.Println("Terraform Version:", state.TerraformVersion)

	if state.Values == nil {
		fmt.Println("No state values found")
		return
	}

	if showOutputs {
		fmt.Println("\nOutputs:")
		if len(state.Values.Outputs) == 0 {
			fmt.Println("  No outputs found")
		}
		for key, output := range state.Values.Outputs {
			fmt.Printf("  %s:\n", key)
			fmt.Printf("    Value: %v\n", output.Value)
			fmt.Printf("    Type: %v\n", output.Type.GoString())
			fmt.Printf("    Sensitive: %v\n", output.Sensitive)
		}
	}

	if showResources {
		fmt.Println("\nResources:")
		if state.Values.RootModule == nil || len(state.Values.RootModule.Resources) == 0 {
			fmt.Println("  No resources found")
			return
		}
		for _, resource := range state.Values.RootModule.Resources {
			fmt.Printf("  Resource: %s\n", resource.Address)
			fmt.Printf("    Type: %s\n", resource.Type)
			fmt.Printf("    Name: %s\n", resource.Name)
			fmt.Printf("    Provider: %s\n", resource.ProviderName)
			fmt.Printf("    Mode: %s\n", resource.Mode)
			if resource.Index != nil {
				fmt.Printf("    Index: %v\n", resource.Index)
			}
			fmt.Printf("    Schema Version: %d\n", resource.SchemaVersion)
			if len(resource.DependsOn) > 0 {
				fmt.Printf("    Depends On: %v\n", resource.DependsOn)
			}
			fmt.Printf("    Tainted: %v\n", resource.Tainted)
			if resource.DeposedKey != "" {
				fmt.Printf("    Deposed Key: %s\n", resource.DeposedKey)
			}
			fmt.Println("    Attributes:")
			for k, v := range resource.AttributeValues {
				fmt.Printf("      %s: %v\n", k, v)
			}
			fmt.Println()
		}
	}
}
