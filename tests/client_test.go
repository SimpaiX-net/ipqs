package tests

import (
	"context"
	"testing"
	"time"

	"github.com/SimpaiX-net/ipqs"
)

func TestClient(t *testing.T) {
	client := ipqs.New()
	client.SetTTL(time.Second * 3)

	err := client.Provision()
	if err != nil {
		t.Fatal(err)
	}

	for _, query := range []string{"1.1.1.1", "0.0.0.0"} {
		if err := client.GetIPQS(context.TODO(), query, "test/bot"); err != nil {
			t.Logf("ipqs: %s", err)

		} else {
			t.Log("ipqs: good")
		}

		time.Sleep(time.Second * 3)

		for range 2 {
			if err := client.GetIPQS(context.TODO(), "1.1.1.1", "test/bot"); err != nil {
				t.Logf("ipqs: %s", err)

			} else {
				t.Log("ipqs: good")
			}
		}
	}
}

func BenchmarkClient(t *testing.B) {
	client := ipqs.New()

	client.SetTTL(time.Millisecond * 200)
	client.Provision()

	t.RunParallel(func(p *testing.PB) {
		for p.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			client.GetIPQS(ctx, "1.1.1.1", "test/bot")

			cancel()
		}
	})
}
