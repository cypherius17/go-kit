package cache

import (
	"context"
	"github.com/1infras/go-kit/src/cmd/cache/redis"
	"github.com/1infras/go-kit/src/cmd/database/elasticsearch"
	rd "github.com/go-redis/redis"
	"testing"
	"time"
)

func TestMultiCache(t *testing.T) {
	r, err := redis.NewDefaultRedisUniversalClient()
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewMultiCache(100, 5 * time.Second, r)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Set("foo", []byte("bar"), 5 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	v, err := c.Get("foo")
	if err != nil {
		t.Fatal(err)
	}

	if p := string(v); p != "bar" {
		t.Fatalf("Expected is bar but actual is %v", p)
	}

	time.Sleep(2 * time.Second)
	v, err = c.Get("foo")

	if p := string(v); p != "bar" {
		t.Fatalf("Expected is bar but actual is %v", p)
	}

	time.Sleep(5 * time.Second)

	v, err = c.Get("foo")
	if err != rd.Nil {
		t.Fatalf("Expected is nil but actual is %v", err)
	}
}

func TestMultiCacheContext(t *testing.T) {
	cd := elasticsearch.DefaultAPMConnection()
	cd.Active = true
	cd.ServiceName = "go-kit"
	cd.Environment = "test"
	cd.AutoBindEnvironment()
	ctx := context.Background()
	r, err := redis.NewDefaultRedisUniversalClient()
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewMultiCache(100, 5 * time.Second, r)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.SetCtx(ctx, "foo", []byte("bar"), 5 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	v, err := c.GetCtx(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if p := string(v); p != "bar" {
		t.Fatalf("Expected is bar but actual is %v", p)
	}

	time.Sleep(2 * time.Second)
	v, err = c.GetCtx(ctx, "foo")

	if p := string(v); p != "bar" {
		t.Fatalf("Expected is bar but actual is %v", p)
	}

	time.Sleep(5 * time.Second)

	v, err = c.GetCtx(ctx, "foo")
	if err != rd.Nil {
		t.Fatalf("Expected is nil but actual is %v", err)
	}
}
