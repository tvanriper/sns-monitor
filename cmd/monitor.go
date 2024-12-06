/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tvanriper/sns-monitor/pkg/monitor"
)

// monitorCmd represents the monitor command
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitors a queue for messages",
	Long: `Monitors a queue for messages, executing an action for any incoming message.
	The body of the message will be passed into the executable as standard input.
	'maxMessage' refers to the maximum number of messages to retrieve at a time.
	'waitSeconds' refers to the time to wait for messages before looping for another try.`,
	Run: func(cmd *cobra.Command, args []string) {
		queue := viper.GetString("queue")
		action := viper.GetString("action")
		maxMessage := viper.GetInt32("maxMessage")
		waitSeconds := viper.GetInt32("waitSeconds")

		actions := strings.Split(action, " ")
		if len(actions) == 0 {
			fmt.Println("must specify an action")
			os.Exit(1)
		}
		command := actions[0]
		arguments := []string{}

		if len(actions) > 1 {
			arguments = actions[1:]
		}

		wg := sync.WaitGroup{}

		wg.Add(1)
		quitting := false
		ctx, err := monitor.SetSQSClientContext(context.Background(), nil)
		if err != nil {
			panic(err)
		}
		go func() {
			defer wg.Done()
			for !quitting {
				iter := monitor.GetMessagesIter(ctx, queue, maxMessage, waitSeconds)
				for msg := range iter {
					fmt.Println("Message arrived.")
					if msg.Body != nil {
						fmt.Println("running:", command)
						fmt.Println("===============")
						cmd := exec.Command(command, arguments...)
						cmd.Stdin = strings.NewReader(*msg.Body)
						cmd.Stdout = os.Stdout
						err := cmd.Run()
						fmt.Println("\n===============")
						if err != nil {
							fmt.Printf("\nfailed to run [%s]: %s\n", action, err)
						} else {
							fmt.Println("\ncommand finished")
						}
						err = monitor.DeleteMessage(ctx, queue, msg)
						if err != nil {
							fmt.Printf("failed to remove message: %s", err)
						}
					}
				}
			}
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		<-c
		quitting = true
		wg.Wait()

	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// monitorCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// monitorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	monitorCmd.PersistentFlags().String("queue", "", "The queue to monitor")
	monitorCmd.PersistentFlags().String("action", "", "The action to run")
	monitorCmd.PersistentFlags().Int16("maxMessage", 10, "Maximum messages to retrieve at a time")
	monitorCmd.PersistentFlags().Int16("waitSeconds", 10, "Maximum time to wait for message in seconds")
	viper.BindPFlag("queue", monitorCmd.PersistentFlags().Lookup("queue"))
	viper.BindPFlag("action", monitorCmd.PersistentFlags().Lookup("action"))
}
