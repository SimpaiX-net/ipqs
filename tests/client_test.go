package tests

import (
	"context"
	"testing"
	"time"

	"github.com/SimpaiX-net/ipqs"
)

func TestCl(t *testing.T) {
	client := ipqs.New()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client.Provision()
	for range 2 {
		client.GetIPQS(ctx, "1.1.1.1", "test/bot")
	}
}
