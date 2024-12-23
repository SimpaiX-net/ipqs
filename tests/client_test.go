package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SimpaiX-net/ipqs"
)

func TestClient(t *testing.T) {
	ipqs.EnableCaching = true

	ctx := context.WithValue(
		context.Background(),
		ipqs.TTL_key,
		time.Hour*24,
	)

	ctx_, cancel := context.WithTimeout(ctx, time.Second * 5)
	defer cancel()
	
	client := ipqs.New()

	err := client.Provision()
	if err != nil {
		t.Fatal(err)
	}

	do := func(query string) {
		start := time.Now()
		err := client.GetIPQS(ctx_, query, "test/bot")
		end := time.Since(start).Milliseconds()

		if err != nil {
			t.Logf("ipqs: %s took %dms", err, end)

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
	ipqs.EnableCaching = true

	ctx := context.WithValue(
		context.Background(),
		ipqs.TTL_key,
		time.Second*2,
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
