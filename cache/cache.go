package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	ctx      context.Context
	db       *redis.Client
	keepTime int64
}

func NewCache(ctx context.Context, db *redis.Client, keepTime int64) *Cache {
	return &Cache{
		ctx:      ctx,
		db:       db,
		keepTime: keepTime,
	}
}

type ICache interface {
	SetUser(msgId int, userId int64) error
	GetUser(msgId int) (int64, error)
	SetBan(userId int64) error
	GetBan(userId int64) (bool, error)
}

func (c *Cache) SetUser(msgId int, userId int64, ) error {
	return c.db.Set(c.ctx, strconv.Itoa(msgId), userId, time.Duration(int64(time.Hour)*c.keepTime)).Err()
}

func (c *Cache) GetUser(msgId int) (int64, error) {
	idStr, err := c.db.Get(c.ctx, strconv.Itoa(msgId)).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(idStr, 10, 64)
}

func (c *Cache) SetBan(userId int64) error {
	idStr := strconv.FormatInt(userId, 10)
	return c.db.Set(c.ctx, idStr, true, time.Hour*100).Err()
}

func (c *Cache) GetBan(userId int64) (bool, error) {
	idStr := strconv.FormatInt(userId, 10)
	return c.db.Get(c.ctx, idStr).Bool()
}
