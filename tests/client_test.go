package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SimpaiX-net/ipqs"
)

var done = make(chan error)

func TestClient(t *testing.T) {
	client := ipqs.New()
	client.SetTTL(time.Second * 3)

	err := client.Provision()
	if err != nil {
		t.Fatal(err)
	}

	do := func(query string) {
		start := time.Now()
		err := client.GetIPQS(context.TODO(), query, "test/bot", done)
		end := time.Since(start).Milliseconds()

		if err != nil {
			t.Logf("ipqs: %s took %dms", err, end)

		} else {
			t.Logf("ipqs: good took %dms", end)
		}

	}
	for _, query := range []string{"1.1.1.1", "0.0.0.0"} {
		fmt.Println("---------new loop----------")
		do(query)

		time.Sleep(time.Second * 4)

		for range 2 {
			do(query)
		}
		fmt.Println("---------end loop----------")
	}
}

func BenchmarkClient(t *testing.B) {
	client := ipqs.New()

	client.SetTTL(time.Millisecond * 200)
	client.Provision()

	t.RunParallel(func(p *testing.PB) {
		for p.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			client.GetIPQS(ctx, "1.1.1.1", "test/bot", done)

			cancel()
		}
	})
}
