package myredis

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWD"), 2)

	b, err := c.AcquireLock("task:001", 60)
	if err != nil {
		t.Error("acquire lock err", err, b)
		return
	}

	//if err := c.ReleaseLock("task:001"); err != nil {
	//	t.Error("release lock err:", err)
	//	return
	//} else {
	//	t.Logf("release lock ok")
	//}

	t.Log("acquire lock result:", b)
}
