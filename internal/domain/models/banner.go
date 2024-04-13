package models

import (
	"time"

	"github.com/AnxVit/avito/internal/domain/models/optional"
)

type BannerDB struct {
	ID      *int64                  `json:"id"`
	Tag     []*int64                `json:"tag_ids"`
	Feature *int64                  `json:"feature_id"`
	Content *map[string]interface{} `json:"content"`
	Access  *bool                   `json:"is_active"`
	Created *time.Time              `json:"created_at"`
	Updated *time.Time              `json:"updated_at"`
}

type BannerPost struct {
	Tag     []int64                `json:"tag_ids"`
	Feature int64                  `json:"feature_id"`
	Content map[string]interface{} `json:"content"`
	Access  bool                   `json:"is_active"`
}

type BannerPatch struct {
	Tag     optional.Optional[[]int64]                `json:"tag_ids"`
	Feature optional.Optional[int64]                  `json:"feature_id"`
	Content optional.Optional[map[string]interface{}] `json:"content"`
	Access  optional.Optional[bool]                   `json:"is_active"`
}
