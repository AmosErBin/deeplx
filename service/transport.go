package service

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func (s *Service) Transport(c *gin.Context) ([]byte, error) {
	endpoint := s.pool.Pick()
	if endpoint == nil {
		return nil, errors.New("not available endpoints")
	}
	if !endpoint.Available() {
		return s.Transport(c)
	}
	req, err := http.NewRequest("POST", endpoint.U, c.Request.Body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(data))
	if gjson.GetBytes(data, "code").Int() != 200 {
		s.pool.Report(endpoint)
		return s.Transport(c)
	}
	return data, nil
}
