package tests

import (
	"context"
	"testing"
	"time"

	"github.com/SimpaiX-net/ipqs"
)

func BenchmarkClient(t *testing.B) {
	client := ipqs.New()

	client.SetTTL(time.Millisecond * 200)

	t.RunParallel(func(p *testing.PB) {
		for p.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

			client.Provision()
			client.GetIPQS(ctx, "1.1.1.1", "test/bot")

			cancel()
		}
	})
}

