package myredis

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type client struct {
	sync.Mutex
	redisClient *redis.Pool
}

func NewClient(address string, password string, db int) Client {
	redisClient := &redis.Pool{
		MaxIdle:     100,
		MaxActive:   500,
		IdleTimeout: time.Second * 5,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", address,
				redis.DialReadTimeout(4*time.Second),
				redis.DialWriteTimeout(4*time.Second),
				redis.DialConnectTimeout(4*time.Second),
				redis.DialDatabase(db),
				redis.DialPassword(password))
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}

	return &client{redisClient: redisClient}
}

func (self *client) Set(key string, value any, ttl int) error {
	bValue, err := self.obj2Stream(value)
	if err != nil {
		return err
	}

	if _, err := self.doCommand("SET", key, bValue); err != nil {
		return err
	}

	if ttl != 0 {
		if _, err := self.doCommand("EXPIRE", key, ttl); err != nil {
			return err
		}
	}

	return nil
}

func (self *client) SetAsJson(key string, v any, ttl int) error {
	bValue, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if _, err := self.doCommand("SET", key, bValue); err != nil {
		return err
	}

	if ttl != 0 {
		if _, err := self.doCommand("EXPIRE", key, ttl); err != nil {
			return err
		}
	}

	return nil
}

func (self *client) Get(key string, model any) error {
	v, err := self.doCommand("GET", key)
	if err != nil {
		return err
	}

	if v == nil {
		return ErrNotFound
	}

	bv, ok := v.([]byte)
	if !ok {
		return fmt.Errorf("incorrect data type")
	}

	return self.stream2Obj(bv, model)
}

func (self *client) GetFromJson(key string, model any) error {
	v, err := self.doCommand("GET", key)
	if err != nil {
		return err
	}

	if v == nil {
		return ErrNotFound
	}

	bv, ok := v.([]byte)
	if !ok {
		return fmt.Errorf("incorrect data type")
	}

	return json.Unmarshal(bv, model)
}

func (self *client) AcquireLock(key string, ttl int) (bool, error) {
	nx, err := redis.Int(self.doCommand("SETNX", key, 1))
	if err != nil {
		return false, err
	}

	if nx == 1 {
		if ttl != 0 {
			if _, err := self.doCommand("EXPIRE", key, ttl); err != nil {
				return false, err
			}
		}
		return true, nil
	}

	return false, nil
}

func (self *client) ReleaseLock(key string) error {
	return self.Del(key)
}

func (self *client) Del(key string) error {
	_, err := self.doCommand("DEL", key)
	return err
}

func (self *client) Expire(key string, ttl int) error {
	_, err := self.doCommand("EXPIRE", key, ttl)
	return err
}

func (self *client) Exist(key string) (bool, error) {
	v, err := self.doCommand("EXISTS", key)
	if err != nil {
		return false, err
	}

	return v.(int64) > 0, nil
}

func (self *client) RemainingTTL(key string) (int64, error) {
	v, err := self.doCommand("TTL", key)
	if err != nil {
		return 0, err
	}

	return v.(int64), nil
}

// ZAdd 向 zset 中添加或更新成员，并可选设置整个 key 的 TTL。
func (self *client) ZAdd(key string, ttl int, items ...ZItem) error {
	if len(items) == 0 {
		return nil
	}

	args := make([]any, 0, 1+len(items)*2)
	args = append(args, key)
	for _, item := range items {
		args = append(args, item.Score, item.Member)
	}

	if _, err := self.doCommand("ZADD", args...); err != nil {
		return err
	}

	if ttl != 0 {
		if _, err := self.doCommand("EXPIRE", key, ttl); err != nil {
			return err
		}
	}

	return nil
}

// ZRem 删除 zset 中的一个或多个成员。
func (self *client) ZRem(key string, members ...any) error {
	if len(members) == 0 {
		return nil
	}

	args := make([]any, 0, 1+len(members))
	args = append(args, key)
	args = append(args, members...)

	_, err := self.doCommand("ZREM", args...)
	return err
}

// ZRemRangeByScore 删除 zset 中指定 score 区间的成员。
func (self *client) ZRemRangeByScore(key, min, max any) error {
	_, err := self.doCommand("ZREMRANGEBYSCORE", key, min, max)
	return err
}

// ZCard 返回 zset 的成员总数。
func (self *client) ZCard(key string) (int64, error) {
	v, err := redis.Int64(self.doCommand("ZCARD", key))
	if err != nil {
		return 0, err
	}
	return v, nil
}

// ZRange 返回 zset 中指定下标范围的成员列表。
func (self *client) ZRange(key string, start, stop int64) ([]string, error) {
	return redis.Strings(self.doCommand("ZRANGE", key, start, stop))
}

// ZRangeByScore 返回 zset 中指定 score 区间的成员列表。
func (self *client) ZRangeByScore(key string, min, max any) ([]string, error) {
	return redis.Strings(self.doCommand("ZRANGEBYSCORE", key, min, max))
}

// ZScore 返回 zset 中指定成员的 score。
func (self *client) ZScore(key string, member any) (float64, error) {
	v, err := redis.Float64(self.doCommand("ZSCORE", key, member))
	if err != nil {
		if err == redis.ErrNil {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return v, nil
}

func (self *client) Conn() error {
	_, err := self.redisClient.Dial()
	return err
}

func (self *client) doCommand(cmd string, arg ...interface{}) (interface{}, error) {
	self.Lock()
	defer self.Unlock()

	conn := self.redisClient.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return nil, err
	}

	return conn.Do(cmd, arg...)
}

func (self *client) obj2Stream(obj any) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(obj); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (self *client) stream2Obj(stream []byte, model any) error {
	dec := gob.NewDecoder(bytes.NewReader(stream))
	return dec.Decode(model)
}
