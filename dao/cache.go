package dao

import (
	"crypto/md5"
	"deeplx/model"
	"encoding/hex"
	"time"

	"github.com/patrickmn/go-cache"
)

type Dao struct {
	cache *cache.Cache
}

func NewDao() *Dao {
	return &Dao{
		cache: cache.New(7*time.Hour, 3*time.Hour),
	}
}

func (d *Dao) SetTransCache(text string, result model.TranslateResp) {
	h := md5.New()
	h.Write([]byte(text))
	key := hex.EncodeToString(h.Sum(nil))
	d.cache.SetDefault(key, result)
}

func (d *Dao) GetTransCache(text string) (model.TranslateResp, bool) {
	h := md5.New()
	h.Write([]byte(text))
	key := hex.EncodeToString(h.Sum(nil))
	var data model.TranslateResp
	v, ok := d.cache.Get(key)
	if !ok {
		return data, false
	}
	return v.(model.TranslateResp), true
}
