/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tvanriper/sns-monitor/pkg/monitor"
)

// queuesCmd represents the queues command
var queuesCmd = &cobra.Command{
	Use:   "queues",
	Short: "Lists SQS queues",
	Long:  `Lists SQS queues from AWS using the default profile.`,
	Run: func(cmd *cobra.Command, args []string) {
		iter := monitor.ListQueuesIter(context.Background())
		for queue := range iter {
			fmt.Println(queue)
		}
	},
}

func init() {
	listCmd.AddCommand(queuesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// queuesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// queuesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
