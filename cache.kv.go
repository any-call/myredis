package myredis

// KVCache 只负责通用缓存：Get / Set / Del
type KVCache[T any] struct {
	client    Client
	namespace string // 业务前缀
	exp       int
}

func NewKVCache[T any](c Client, namespace string, ttlSec int) *KVCache[T] {
	if c == nil {
		panic("client is nil")
	}

	if namespace == "" {
		panic("key is empty")
	}

	return &KVCache[T]{client: c, namespace: namespace, exp: ttlSec}
}

// 拼接最终存储 key
func (c *KVCache[T]) buildKey(key string) string {
	return c.namespace + "::" + key
}

func (c *KVCache[T]) Set(key string, value T) error {
	ttl := c.exp
	return c.client.Set(c.buildKey(key), value, ttl)
}

func (c *KVCache[T]) Get(key string) (T, error) {
	var ret T
	err := c.client.Get(c.buildKey(key), &ret)
	return ret, err
}

func (c *KVCache[T]) Del(key string) error {
	return c.client.Del(c.buildKey(key))
}

func (c *KVCache[T]) Expire(key string, exp ...int) error {
	ttl := c.exp
	if len(exp) > 0 {
		ttl = exp[0]
	}
	return c.client.Expire(c.buildKey(key), ttl)
}
