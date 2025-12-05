package myredis

import (
	"fmt"
	"sort"
)

type DBCache[T any] struct {
	client Client
	key    string              // Redis key
	loadFn func() ([]T, error) // DB 加载接口
	exp    int                 // 过期时间（可选）
}

func NewDBCache[T any](c Client, key string, loadFn func() ([]T, error), ttl int) *DBCache[T] {
	if c == nil {
		panic("client is nil")
	}

	if key == "" {
		panic("key is empty")
	}

	if loadFn == nil {
		panic("load fn is nil ")
	}

	return &DBCache[T]{
		client: c,
		key:    key,
		loadFn: loadFn,
		exp:    ttl,
	}
}

func (r *DBCache[T]) List() ([]T, error) {
	var list []T
	err := r.client.GetFromJson(r.key, &list)
	if err == nil {
		return list, nil
	}

	// 缓存未命中 → DB 查询并更新
	return r.refresh()
}

func (r *DBCache[T]) First(condition func(ret T) bool) (*T, error) {
	list, err := r.List()
	if err != nil {
		return nil, err
	}
	if condition == nil {
		if len(list) > 0 {
			return &list[0], err
		}
		return nil, fmt.Errorf("empty record")
	}

	for _, item := range list {
		if condition(item) {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("no match record")
}

func (r *DBCache[T]) Find(cond func(T) bool) ([]T, error) {
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

func (r *DBCache[T]) FindSorted(
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

func (r *DBCache[T]) Invalidate() error {
	return r.client.Del(r.key)
}

func (r *DBCache[T]) refresh() ([]T, error) {
	list, err := r.loadFn()
	if err != nil {
		return nil, err
	}

	if err := r.client.SetAsJson(r.key, list, r.exp); err != nil {
		return nil, err
	}

	return list, nil
}
