package myredis

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

const (
	OneSecond = 1
	OneMinute = 60 * OneSecond
	OneHour   = 60 * OneMinute
	OneDay    = OneHour * 24
)

type Client interface {
	Set(key string, v any, ttl int) error //ttl==0 ，永不过期
	SetAsJson(key string, v any, ttl int) error
	AcquireLock(key string, ttl int) (bool, error)
	ReleaseLock(key string) error
	Get(key string, v any) error
	GetFromJson(key string, v any) error
	Del(key string) error
	Exist(key string) (bool, error)
	RemainingTTL(key string) (int64, error)
	Expire(key string, ttl int) error
	Conn() error
}

var (
	ErrNotFound = fmt.Errorf("the key not found")
)

func StreamToObject[E any](b []byte) (ret E, err error) {
	reader := bytes.NewReader(b)
	dec := gob.NewDecoder(reader)
	if err = dec.Decode(&ret); err != nil {
		return ret, err
	}

	return ret, nil
}

//redis 常用指令说明
//1. SET：设置一个键值对。
//2. GET：获取指定键的值。
//3. DEL：删除指定键。
//4. EXISTS：检查指定键是否存在。
//5. KEYS：获取匹配指定模式的键列表。
//6. EXPIRE：设置键的过期时间。
//7. TTL：获取键的剩余过期时间。
//8. INCR：将键的值增加1。
//9. DECR：将键的值减少1。
//10. HSET：在哈希类型键中设置字段和值。
//11. HGET：获取哈希类型键中指定字段的值。
//12. HGETALL：获取哈希类型键中所有字段和值的列表。
//13. LPUSH：将值从列表的左侧插入。
//14. RPUSH：将值从列表的右侧插入。
//15. LPOP：从列表的左侧弹出一个值。
//16. RPOP：从列表的右侧弹出一个值。
