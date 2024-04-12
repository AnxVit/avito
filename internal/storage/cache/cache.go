package cache

import (
	"Auth-Reg/internal/storage/cache/debounce"
	"Auth-Reg/internal/storage/postgres"
	"sync"
	"time"
)

type Cache struct {
	DB       *postgres.Db
	cache    sync.Map
	debounce map[[2]int]interface{}
}

func New(db *postgres.Db) (*Cache, error) {
	return &Cache{
		DB:       db,
		debounce: make(map[[2]int]interface{}),
	}, nil
}

func (c *Cache) GetUserBanner(tag, feature int, use_last_reversion bool, admin bool) (map[string]interface{}, error) {
	key := [2]int{tag, feature}

	if use_last_reversion {
		return c.DB.GetUserBanner(tag, feature, admin)
	}

	banner_int, ok := c.cache.Load(key)

	if !ok {
		banner, err := c.DB.GetUserBanner(tag, feature, admin)
		if err != nil {
			return nil, err
		}
		c.cache.Store(key, banner)
		if _, ok := c.debounce[key]; !ok {
			c.debounce[key] = debounce.New(5 * time.Minute)
		}

		c.debounce[key].(func(f func()))(func() {
			c.cache.Delete(key)
		})
		return banner, nil
	}
	c.debounce[key].(func(f func()))(func() {
		c.cache.Delete(key)
	})
	return banner_int.(map[string]interface{}), nil
}
