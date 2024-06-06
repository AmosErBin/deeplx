package service

import (
	"bytes"
	"deeplx/dao"
	"deeplx/model"
	"deeplx/pool"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

type Service struct {
	cLock sync.RWMutex
	dao   *dao.Dao
	pool  *pool.Pool
}

func NewService() *Service {
	return &Service{
		cLock: sync.RWMutex{},
		dao:   dao.NewDao(),
		pool:  pool.InitPool(),
	}
}

func (s *Service) Translate(text, sourceLang, targetLang, cookie string) (*model.TranslateResp, error) {
	v, ok := s.dao.GetTransCache(fmt.Sprintf("%s_%s_%s", text, sourceLang, targetLang))
	if ok {
		fmt.Println("hit cache")
		return &v, nil
	}
	id := getRandomNumber()
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "LMT_handle_texts",
		"id":      id,
		"params": map[string]interface{}{
			"texts": []map[string]interface{}{{
				"text":                text,
				"requestAlternatives": 3,
			}},
			"splitting": "newlines",
			"lang": map[string]interface{}{
				"source_lang_user_selected": sourceLang,
				"target_lang":               targetLang,
			},
			"timestamp": getTimestamp(int64(getICount(text))),
			"common_job_params": map[string]interface{}{
				"was_spoken":    false,
				"transcribe_as": "",
			},
		},
	}
	//body转换io.Reader
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.deepl.com/jsonrpc", bytes.NewReader(data))
	for k, v := range model.HEADERS {
		req.Header.Add(k, v)
	}
	req.Header.Add("cookie", "dl_session="+cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	var alternatives []string
	gjson.GetBytes(result, "result.texts.0.alternatives").ForEach(func(key, value gjson.Result) bool {
		alternatives = append(alternatives, value.Get("text").String())
		return true
	})

	text = gjson.GetBytes(result, "result.texts.0.text").String()
	r := &model.TranslateResp{
		Code:         200,
		Data:         gjson.GetBytes(result, "result.texts.0.text").String(),
		Id:           id,
		Method:       "Free",
		SourceLang:   sourceLang,
		TargetLang:   targetLang,
		Alternatives: alternatives,
	}
	if text == "" {
		fmt.Println(string(result))
		r.Code = 500
		return r, nil
	}
	s.dao.SetTransCache(fmt.Sprintf("%s_%s_%s", text, sourceLang, targetLang), *r)
	return r, nil
}

func getICount(text string) int {
	return len(strings.Split(text, "i"))
}

func getRandomNumber() int64 {
	return (rand.Int63n(99999) + 8300000) * 1000
}

func getTimestamp(iCount int64) int64 {
	ts := time.Now().UnixMilli()
	if iCount != 0 {
		iCount = iCount + 1
		return ts - ts%iCount + iCount
	} else {
		return ts
	}
}
