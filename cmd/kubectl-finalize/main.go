package main

import (
    "fmt"
    "os"

    "kubectl-finalize/pkg/rm"
    "github.com/spf13/cobra"
    "k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
    streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
    cmd := NewRootCmd(streams)
    
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func NewRootCmd(streams genericclioptions.IOStreams) *cobra.Command {
    flags := genericclioptions.NewConfigFlags(true)
    
    cmd := &cobra.Command{
        Use:   "kubectl-finalize RESOURCE",
        Short: "Force delete Kubernetes resources stuck in Terminating state",
        Long: `A kubectl plugin to force delete Kubernetes resources that are stuck in Terminating state.
It removes finalizers and performs a force deletion of the resource.`,
        Example: `  # Force delete a pod
  kubectl finalize pod/stuck-pod

  # Force delete a namespace
  kubectl finalize namespace/stuck-ns

  # Force delete a resource in specific namespace
  kubectl finalize deployment/stuck-deploy -n my-namespace`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return rm.ForceDelete(flags, streams, args)
        },
    }

    flags.AddFlags(cmd.Flags())
    return cmd
} 
