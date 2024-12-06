package monitor

import (
	"context"
	"fmt"
	"iter"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snsTypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func readyCtxConfig(ctx context.Context) (result context.Context, err error) {
	result = ctx
	if ctx.Value(ctxConfig) == nil {
		var sdkConfig aws.Config
		sdkConfig, err = config.LoadDefaultConfig(ctx)
		if err != nil {
			slog.Error("unable to get client from default AWS configuration", "error", err)
			return
		}
		result = context.WithValue(ctx, ctxConfig, sdkConfig)
	}
	return
}

// getCtxConfig gets the aws.Config within the context, or returns an error.
// It will attempt to get the default configuration if the context isn't configured.
func getCtxConfig(ctx context.Context) (result aws.Config, err error) {
	var ok bool
	if result, ok = ctx.Value(ctxConfig).(aws.Config); !ok {
		var newCtx context.Context
		newCtx, err = readyCtxConfig(ctx)
		if err != nil {
			return
		}
		if result, ok = newCtx.Value(ctxConfig).(aws.Config); !ok {
			err = fmt.Errorf("unable to get default AWS configuration")
		}
	}
	return
}

// SetCtxConfig sets the aws.Config for this context.
func SetCtxConfig(ctx context.Context, conf aws.Config) (result context.Context) {
	result = context.WithValue(ctx, ctxConfig, conf)
	return
}

// SetCtxDefaultConfig sets the aws.Config for this context to the AWS default configuration.
func SetCtxDefaultConfig(ctx context.Context) (result context.Context, err error) {
	return readyCtxConfig(ctx)
}

// readySNSClient fetches the client for sns calls.
func readySNSClient(ctx context.Context) (result context.Context, err error) {
	result = ctx
	if ctx.Value(ctxSNSClient) == nil {
		var sdkConfig aws.Config
		sdkConfig, err = getCtxConfig(ctx)
		if err != nil {
			slog.Error("unable to get client from default AWS configuration", "error", err)
			return
		}
		client := sns.NewFromConfig(sdkConfig)
		result = context.WithValue(ctx, ctxSNSClient, client)
	}
	return
}

// getSNSClient retrieves the client from the context, or attempts to get a default client using readySNSClient.
func getSNSClient(ctx context.Context) (result *sns.Client) {
	var ok bool
	result, ok = ctx.Value(ctxSNSClient).(*sns.Client)
	if !ok {
		// Try again with readySNSClient
		myContext, err := readySNSClient(ctx)
		if err == nil {
			result, ok = myContext.Value(ctxSNSClient).(*sns.Client)
			if ok {
				// We have a client.
				return
			}
		}
		result = nil
	}
	return
}

// getSQSClient retrieves the client from the context, or attempts to get a default client using readySQSClient.
func getSQSClient(ctx context.Context) (result *sqs.Client) {
	var ok bool
	result, ok = ctx.Value(ctxSQSClient).(*sqs.Client)
	if !ok {
		myContext, err := readySQSClient(ctx)
		if err == nil {
			result, ok = myContext.Value(ctxSQSClient).(*sqs.Client)
			if ok {
				return
			}
		}
		result = nil
	}
	return
}

// readySQSClient retrieves a context with an SQS client, even if it's from the default configuration.
func readySQSClient(ctx context.Context) (result context.Context, err error) {
	result = ctx
	if ctx.Value(ctxSQSClient) == nil {
		var sdkConfig aws.Config
		sdkConfig, err = getCtxConfig(ctx)
		if err != nil {
			slog.Error("unable to get client from default AWS configuration", "error", err)
			return
		}
		if client := sqs.NewFromConfig(sdkConfig); client != nil {
			result = context.WithValue(ctx, ctxSQSClient, client)
		} else {
			err = fmt.Errorf("unable to create SQS client")
		}
	}
	return
}

// SetSQSClientContext provides a context with the desired client to use for SQS requests.
// Or, if you pass in a nil client, it will try to create a client from your default AWS settings.
func SetSQSClientContext(ctx context.Context, client *sqs.Client) (result context.Context, err error) {
	if client == nil {
		result, err = readySQSClient(ctx)
		return
	}
	result = context.WithValue(ctx, ctxSQSClient, client)
	return
}

// SetSNSClientContext provides a context with the desired client to use for SNS requests.
// Or if you pass in a nil client, it will try to create a client from your default AWS settings.
func SetSNSClientContext(ctx context.Context, client *sns.Client) (result context.Context, err error) {
	if client == nil {
		result, err = readySNSClient(ctx)
		return
	}
	result = context.WithValue(ctx, ctxSNSClient, client)
	return
}

// ListTopicsIter provides an iterator to a fetched collection of snsTypes.Topic from AWS.
func ListTopicsIter(ctx context.Context) (result iter.Seq[snsTypes.Topic]) {
	return func(yield func(topic snsTypes.Topic) bool) {
		client := getSNSClient(ctx)
		if client == nil {
			slog.Warn("unable to list topics without an SNS client")
			return
		}
		paginator := sns.NewListTopicsPaginator(client, &sns.ListTopicsInput{})
		for paginator.HasMorePages() {
			output, err := paginator.NextPage(ctx)
			if err != nil {
				slog.Warn("unable to get next page of topics", "error", err)
				return
			}
			for _, topic := range output.Topics {
				if !yield(topic) {
					return
				}
			}
		}
	}
}

// ListQueuesIter returns a list of SQS queue urls that are available.
func ListQueuesIter(ctx context.Context) (result iter.Seq[string]) {
	return func(yield func(q string) bool) {
		client := getSQSClient(ctx)
		paginator := sqs.NewListQueuesPaginator(client, &sqs.ListQueuesInput{})
		for paginator.HasMorePages() {
			output, err := paginator.NextPage(ctx)
			if err != nil {
				slog.Debug("could not get queues", "error", err)
				return
			}
			if output == nil {
				return
			}
			for _, item := range output.QueueUrls {
				if !yield(item) {
					return
				}
			}
		}
	}
}

// GetMessagesIter provides and iterator to messages in a queue.
// It expects the context to have an SQS client, or it will create one from the default configuration if needed.
func GetMessagesIter(ctx context.Context, queue string, maxMessage int32, waitSeconds int32) (result iter.Seq[sqsTypes.Message]) {
	return func(yield func(msg sqsTypes.Message) bool) {
		client := getSQSClient(ctx)
		if client == nil {
			return
		}
		messages, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queue),
			MaxNumberOfMessages: maxMessage,
			WaitTimeSeconds:     waitSeconds,
		})
		if err != nil {
			slog.Warn("failed to fetch messages", "error", err)
			return
		}
		for _, message := range messages.Messages {
			if !yield(message) {
				return
			}
		}
	}
}

// DeleteMessage removes a message from the queue.
func DeleteMessage(ctx context.Context, queue string, message sqsTypes.Message) (err error) {
	client := getSQSClient(ctx)
	if client == nil {
		err = fmt.Errorf("unable to acquire SQS client from context")
		return
	}
	_, err = client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queue),
		ReceiptHandle: message.ReceiptHandle,
	})
	return
}
