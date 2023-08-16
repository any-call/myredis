package myredis

import (
	"os"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c := NewClient(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWD"), 3)
	if err := c.Set("001", "001", 5); err != nil {
		t.Error("set err", err)
		return
	}

	if v, err := c.Exist("001"); err != nil {
		t.Error("exist err:", err)
		return
	} else {
		t.Log("exist 001:", v)
	}

	if v, err := c.RemainingTTL("001"); err != nil {
		t.Error("remaining err:", err)
		return
	} else {
		t.Log("remaining 001:", v)
	}

	if v, err := c.Get("001"); err != nil {
		t.Error("1 get err:", err)
		return
	} else {
		t.Log("get 001:", v)
		a, ok := To[string](v)
		t.Log("get 001:", a, ok)
	}

	time.Sleep(time.Second * 6)

	if v, err := c.Get("001"); err != nil {
		t.Error("2 get err:", err)
		return
	} else {
		t.Log("get 001", v)
	}

	t.Log("test ok")
}
