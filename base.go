package myredis

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
)

var ErrRecordNotFound = errors.New("record not found")

const (
	OneSecond = 1
	OneMinute = 60 * OneSecond
	OneHour   = 60 * OneMinute
	OneDay    = OneHour * 24
)

type ZItem struct {
	Score  float64
	Member any
}

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

	//增加Zset 标准封装
	// ZAdd 向有序集合添加一个或多个成员。
	// 如果成员已存在，则更新其 score。
	// ttl==0 表示不设置过期时间，否则写入后会给整个 key 设置过期时间。
	ZAdd(key string, ttl int, items ...ZItem) error
	// ZRem 从有序集合中删除一个或多个成员。
	// 不存在的成员会被忽略。
	ZRem(key string, members ...any) error

	// ZRemRangeByScore 删除指定 score 区间内的所有成员。
	// 常用于清理“过期时间作为 score”的数据。
	// min / max 可传数字，也可传 Redis 支持的区间表达式，如 "-inf"、"+inf"。
	ZRemRangeByScore(key, min, max any) error

	// ZCard 返回有序集合当前成员数量。
	ZCard(key string) (int64, error)

	// ZRange 按 score 升序返回指定下标范围内的成员。
	// start / stop 的语义与 Redis ZRANGE 一致，支持负数下标。
	ZRange(key string, start, stop int64) ([]string, error)

	// ZRangeByScore 按 score 区间返回成员列表，结果按 score 升序排列。
	// min / max 可传数字，也可传 Redis 支持的区间表达式，如 "-inf"、"+inf"。
	ZRangeByScore(key string, min, max any) ([]string, error)

	// ZScore 查询指定成员当前的 score。
	// 如果成员不存在，返回 ErrNotFound。
	ZScore(key string, member any) (float64, error)

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
