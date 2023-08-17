package myredis

import (
	"os"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c := NewClient(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWD"), 3)
	type Tmpstuct struct {
		ID   int64
		Name string
	}

	type Tmpstuct11 struct {
		ID   int64
		Name string
	}

	if err := c.Set("001", "4232323", 5); err != nil {
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
		ret, err := StreamToObject[string](v)
		t.Log("get 001:", ret, err)
	}

	time.Sleep(time.Second * 3)
	t.Log("after time.sleep: 6s ")
	if v, err := c.Exist("001"); err != nil {
		t.Error("exist err:", err)
		return
	} else {
		t.Log("exist 001", v)
	}

	if v, err := c.RemainingTTL("001"); err != nil {
		t.Error("2 get err:", err)
		return
	} else {
		t.Log("remain ttl 001", v)
	}

	if err := c.Expire("001", 5); err != nil {
		t.Error("expire err:", err)
	}

	if v, err := c.RemainingTTL("001"); err != nil {
		t.Error("3 get err:", err)
		return
	} else {
		t.Log("3  remain ttl 001", v)
	}

	//if err := c.Del("001"); err != nil {
	//	t.Error("3 delete err:", err)
	//	return
	//}

	if v, err := c.Get("001"); err != nil {
		t.Error("1 get err:", err)
		return
	} else {
		ret, err := StreamToObject[map[string]any](v)
		t.Log("get 001:", ret, err)
	}

	t.Log("test ok")
}
