package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SimpaiX-net/ipqs"
)

func TestClient(t *testing.T) {
	ctx := context.WithValue(
		context.Background(),
		ipqs.TTL_key,
		time.Second*5,
	)
	client := ipqs.New()

	err := client.Provision()
	if err != nil {
		t.Fatal(err)
	}

	do := func(query string) {
		start := time.Now()
		err := client.GetIPQS(ctx, query, "test/bot")
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

		time.Sleep(time.Second * 1)

		for range 2 {
			do(query)
		}
		fmt.Println("---------end loop----------")
	}
}

func BenchmarkClient(t *testing.B) {
	ctx := context.WithValue(
		context.Background(),
		ipqs.TTL_key,
		time.Second*1,
	)

	client := ipqs.New()
	client.Provision()

	t.RunParallel(func(p *testing.PB) {
		for p.Next() {
			ctx_, cancel := context.WithTimeout(ctx, time.Millisecond*200)
			client.GetIPQS(ctx_, "1.1.1.1", "test/bot")

			cancel()
		}
	})
}
