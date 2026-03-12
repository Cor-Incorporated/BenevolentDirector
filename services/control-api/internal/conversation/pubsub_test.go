package conversation

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestPubSubPublisher_NilPublisher(t *testing.T) {
	var pub *PubSubPublisher
	err := pub.Publish(context.Background(), "topic", "key", []byte("data"))
	if err == nil {
		t.Fatal("expected error for nil publisher")
	}
}

func TestPubSubPublisher_NilClient(t *testing.T) {
	pub := &PubSubPublisher{client: nil}
	err := pub.Publish(context.Background(), "topic", "key", []byte("data"))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewPubSubPublisher(t *testing.T) {
	pub := NewPubSubPublisher(nil)
	if pub == nil {
		t.Fatal("NewPubSubPublisher returned nil")
	}
}

func mustTestPubSubClient(t *testing.T) *pubsub.Client {
	t.Helper()
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "test-project",
		option.WithEndpoint("localhost:1"),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		t.Fatalf("failed to create test pubsub client: %v", err)
	}
	return client
}

func TestPubSubPublisher_EmptyTopic(t *testing.T) {
	client := mustTestPubSubClient(t)
	defer client.Close()

	pub := NewPubSubPublisher(client)
	err := pub.Publish(context.Background(), "", "key", []byte("data"))
	if err == nil {
		t.Fatal("expected error for empty topic")
	}
}

func TestPubSubPublisher_EmptyOrderingKey(t *testing.T) {
	client := mustTestPubSubClient(t)
	defer client.Close()

	pub := NewPubSubPublisher(client)
	err := pub.Publish(context.Background(), "topic", "", []byte("data"))
	if err == nil {
		t.Fatal("expected error for empty ordering key")
	}
}
