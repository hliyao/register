package grpclb

import (
	"testing"
	"time"

	"log"

	"golang.org/x/net/context"
)

func init() {
}

var r = DefaultEtcdv3Resolver()

func TestEtcdv3Resolver_Register(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	target := "test1.52tt.local"
	r.Register(ctx, target, "localhost:8080")

	select {
	case <-ctx.Done():
		log.Println("Context done:", ctx.Err())
		break
	}
	//time.Sleep(500*time.Second)
}

func TestEtcdv3Resolver_Unregister(t *testing.T) {
	r.Unregister(context.Background(), "test.52tt.local", "localhost:8080")
}