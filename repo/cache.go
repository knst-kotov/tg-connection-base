package repo

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	db       *redis.Client
	keepTime int64
}

func NewCache(db *redis.Client, keepTime int64) *Cache {
	return &Cache{
		db:       db,
		keepTime: keepTime,
	}
}

type ICache interface {
	SetUser(ctx context.Context, userId int64, msgId int) error
	GetUser(ctx context.Context, msgId int) (int64, error)
	SetBan(ctx context.Context, userId int64) error
	GetBan(ctx context.Context, userId int64) (bool, error)
}

func (c *Cache) SetUser(ctx context.Context, userId int64, msgId int) error {
	return c.db.Set(ctx, strconv.Itoa(msgId), userId, time.Duration(int64(time.Hour)*c.keepTime)).Err()
}

func (c *Cache) GetUser(ctx context.Context, msgId int) (int64, error) {
	idStr, err := c.db.Get(ctx, strconv.Itoa(msgId)).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(idStr, 10, 64)
}

func (c *Cache) SetBan(ctx context.Context, userId int64) error {
	idStr := strconv.FormatInt(userId, 10)
	return c.db.Set(ctx, idStr, true, time.Hour*100).Err()
}

func (c *Cache) GetBan(ctx context.Context, userId int64) (bool, error) {
	idStr := strconv.FormatInt(userId, 10)
	return c.db.Get(ctx, idStr).Bool()
}
