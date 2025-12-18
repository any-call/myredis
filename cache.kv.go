package myredis

import "fmt"

// KVCache 只负责通用缓存：Get / Set / Del
type KVCache[K comparable, V any] struct {
	client    Client
	namespace string // 业务前缀
	exp       int
}

func NewKVCache[K comparable, V any](c Client, namespace string, ttlSec int) *KVCache[K, V] {
	if c == nil {
		panic("client is nil")
	}

	if namespace == "" {
		panic("key is empty")
	}

	return &KVCache[K, V]{client: c, namespace: namespace, exp: ttlSec}
}

// 拼接最终存储 key
func (c *KVCache[K, V]) buildKey(key K) string {
	return fmt.Sprintf("%s::%v", c.namespace, key)
}

func (c *KVCache[K, V]) Set(key K, value V) error {
	ttl := c.exp
	return c.client.Set(c.buildKey(key), value, ttl)
}

func (c *KVCache[K, V]) Get(key K) (V, error) {
	var ret V
	err := c.client.Get(c.buildKey(key), &ret)
	return ret, err
}

func (c *KVCache[K, V]) Del(key K) error {
	return c.client.Del(c.buildKey(key))
}

func (c *KVCache[K, V]) Expire(key K, exp ...int) error {
	ttl := c.exp
	if len(exp) > 0 {
		ttl = exp[0]
	}
	return c.client.Expire(c.buildKey(key), ttl)
}
