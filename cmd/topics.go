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

// topicsCmd represents the topic command
var topicsCmd = &cobra.Command{
	Use:   "topics",
	Short: "Lists SNS topics in your default profile",
	Long:  `Lists the SNS topics available to your default profile`,
	Run: func(cmd *cobra.Command, args []string) {
		iter := monitor.ListTopicsIter(context.Background())
		for topic := range iter {
			if topic.TopicArn != nil {
				fmt.Println(*topic.TopicArn)
			}
		}
	},
}

func init() {
	listCmd.AddCommand(topicsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topicCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// topicCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
