package myredis

import (
	"bytes"
	"encoding/gob"
	"github.com/gomodule/redigo/redis"
	"sync"
	"time"
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

func (self *client) Get(key string) ([]byte, error) {
	v, err := self.doCommand("GET", key)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, ErrNotFound
	}

	return v.([]byte), nil
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
