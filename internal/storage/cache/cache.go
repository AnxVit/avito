package cache

import (
	"sync"
	"time"

	"github.com/AnxVit/avito/internal/storage/cache/debounce"
	"github.com/AnxVit/avito/internal/storage/postgres"
)

type Cache struct {
	DB       *postgres.Repo
	cache    sync.Map
	debounce map[[2]int]func(f func())
}

func New(db *postgres.Repo) (*Cache, error) {
	return &Cache{
		DB:       db,
		debounce: make(map[[2]int]func(f func())),
	}, nil
}

func (c *Cache) GetUserBanner(tag, feature int, useLastReversion bool, admin bool) (map[string]interface{}, error) {
	key := [2]int{tag, feature}

	if useLastReversion {
		return c.DB.GetUserBanner(tag, feature, admin)
	}

	bannerInterface, ok := c.cache.Load(key)

	if !ok {
		banner, err := c.DB.GetUserBanner(tag, feature, admin)
		if err != nil {
			return nil, err
		}
		c.cache.Store(key, banner)
		if _, ok := c.debounce[key]; !ok {
			c.debounce[key] = debounce.New(5 * time.Minute)
		}
		c.debounce[key](func() {
			c.cache.Delete(key)
		})
		return banner, nil
	}
	c.debounce[key](func() {
		c.cache.Delete(key)
	})
	return bannerInterface.(map[string]interface{}), nil //nolint:forcetypeassert
}
