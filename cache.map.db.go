package myredis

type DBMapCache[K comparable, V any] struct {
	client Client
	key    string                       // Redis key
	loadFn func() (map[K]V, int, error) // DB 加载接口
}

func NewDBMapCache[K comparable, V any](c Client, key string, loadFn func() (map[K]V, int, error)) *DBMapCache[K, V] {
	if c == nil {
		panic("client is nil")
	}

	if key == "" {
		panic("key is empty")
	}

	if loadFn == nil {
		panic("load fn is nil ")
	}

	return &DBMapCache[K, V]{
		client: c,
		key:    key,
		loadFn: loadFn,
	}
}

func (r *DBMapCache[K, V]) Map() (map[K]V, error) {
	var list map[K]V
	err := r.client.GetFromJson(r.key, &list)
	if err == nil {
		return list, nil
	}

	// 缓存未命中 → DB 查询并更新
	return r.refresh()
}

func (r *DBMapCache[K, V]) Get(key K) (*V, error) {
	m, err := r.Map()
	if err != nil {
		return nil, err
	}

	if v, ok := m[key]; ok {
		return &v, nil
	}

	return nil, ErrRecordNotFound
}

func (r *DBMapCache[K, V]) Invalidate() error {
	return r.client.Del(r.key)
}

func (r *DBMapCache[K, V]) refresh() (map[K]V, error) {
	list, ttl, err := r.loadFn()
	if err != nil {
		return nil, err
	}

	if err := r.client.SetAsJson(r.key, list, ttl); err != nil {
		return nil, err
	}

	return list, nil
}
