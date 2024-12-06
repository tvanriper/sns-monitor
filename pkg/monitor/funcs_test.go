package monitor_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/tvanriper/sns-monitor/pkg/monitor"
)

func TestListTopicsIter(t *testing.T) {
	// NOTE:
	// This test requires that you set up your default AWS configuration
	// appropriately to work properly.  Sorry that it isn't more automated.
	// Maybe if someone goes to the trouble to create a way to mock AWS
	// requests... (tongue firmly in cheek on that expectation)
	ctx, err := monitor.SetSNSClientContext(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	iter := monitor.ListTopicsIter(ctx)
	for item := range iter {
		if item.TopicArn != nil {
			t.Logf(*item.TopicArn)
		}
	}
}

func TestListQueues(t *testing.T) {
	ctx, err := monitor.SetSQSClientContext(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	iter := monitor.ListQueuesIter(ctx)
	for item := range iter {
		t.Logf(item)
	}
}

func TestGetMessages(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx, err := monitor.SetSQSClientContext(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	iter := monitor.ListQueuesIter(ctx)
	for item := range iter {
		msgs := monitor.GetMessagesIter(ctx, item, 10, 12)
		for msg := range msgs {
			t.Logf("received message: %#v", msg)
			if msg.MessageId != nil {
				t.Logf("MessageId: %s", *msg.MessageId)
			}
			if msg.Body != nil {
				t.Logf("Body: %s", *msg.Body)
			}
			err = monitor.DeleteMessage(ctx, item, msg)
			if err != nil {
				t.Error(err)
			}
		}
	}
}
