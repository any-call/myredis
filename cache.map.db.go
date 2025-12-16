package myredis

import "sync"

type DBMapCache[K comparable, V any] struct {
	sync.Mutex // 新增
	client     Client
	key        string                       // Redis key
	loadFn     func() (map[K]V, int, error) // DB 加载接口
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

	// 2️⃣ 加锁，防止并发 refresh
	r.Lock()
	defer r.Unlock()

	// 3️⃣ double check（非常关键）
	if err := r.client.GetFromJson(r.key, &list); err == nil {
		return list, nil
	}

	// 4️⃣ 真正 refresh
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

func (r *DBMapCache[K, V]) First(condition func(ret V) bool) (*V, error) {
	list, err := r.Map()
	if err != nil {
		return nil, err
	}

	for _, item := range list {
		if condition == nil {
			return &item, nil
		} else {
			if condition(item) {
				return &item, nil
			}
		}
	}

	return nil, ErrRecordNotFound
}

func (r *DBMapCache[K, V]) Find(cond func(V) bool) ([]V, error) {
	list, err := r.Map()
	if err != nil {
		return nil, err
	}

	var ret []V
	for _, item := range list {
		if cond == nil {
			ret = append(ret, item)
		} else {
			if cond(item) {
				ret = append(ret, item)
			}
		}
	}

	return ret, nil
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
