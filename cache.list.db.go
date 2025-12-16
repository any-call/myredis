package myredis

import (
	"sort"
	"sync"
)

type DBListCache[T any] struct {
	sync.Mutex // 新增
	client     Client
	key        string                   // Redis key
	loadFn     func() ([]T, int, error) // DB 加载接口
}

func NewDBListCache[T any](c Client, key string, loadFn func() ([]T, int, error)) *DBListCache[T] {
	if c == nil {
		panic("client is nil")
	}

	if key == "" {
		panic("key is empty")
	}

	if loadFn == nil {
		panic("load fn is nil ")
	}

	return &DBListCache[T]{
		client: c,
		key:    key,
		loadFn: loadFn,
	}
}

func (r *DBListCache[T]) List() ([]T, error) {
	var list []T
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

func (r *DBListCache[T]) First(condition func(ret T) bool) (*T, error) {
	list, err := r.List()
	if err != nil {
		return nil, err
	}
	if condition == nil {
		if len(list) > 0 {
			return &list[0], err
		}
		return nil, ErrRecordNotFound
	}

	for _, item := range list {
		if condition(item) {
			return &item, nil
		}
	}

	return nil, ErrRecordNotFound
}

func (r *DBListCache[T]) FirstSorted(
	cond func(T) bool,
	less func(a, b T) bool,
) (*T, error) {
	list, err := r.Find(cond)
	if err != nil {
		return nil, err
	}

	if list == nil || len(list) == 0 {
		return nil, ErrRecordNotFound
	}

	if less != nil {
		sort.Slice(list, func(i, j int) bool {
			return less(list[i], list[j])
		})
	}

	return &list[0], nil
}

func (r *DBListCache[T]) Find(cond func(T) bool) ([]T, error) {
	list, err := r.List()
	if err != nil {
		return nil, err
	}
	if cond == nil {
		return list, nil
	}
	var ret []T
	for _, item := range list {
		if cond(item) {
			ret = append(ret, item)
		}
	}
	return ret, nil
}

func (r *DBListCache[T]) FindSorted(
	cond func(T) bool,
	less func(a, b T) bool,
) ([]T, error) {
	list, err := r.Find(cond)
	if err != nil {
		return nil, err
	}

	if less != nil {
		sort.Slice(list, func(i, j int) bool {
			return less(list[i], list[j])
		})
	}

	return list, nil
}

func (r *DBListCache[T]) Invalidate() error {
	return r.client.Del(r.key)
}

func (r *DBListCache[T]) refresh() ([]T, error) {
	list, ttl, err := r.loadFn()
	if err != nil {
		return nil, err
	}

	if err := r.client.SetAsJson(r.key, list, ttl); err != nil {
		return nil, err
	}

	return list, nil
}
