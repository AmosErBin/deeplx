package pool

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tidwall/gjson"
)

type Pool struct {
	lock sync.RWMutex

	index atomic.Int64

	head []*Endpoints
}

type Endpoints struct {
	U string
	A atomic.Bool
}

func (e *Endpoints) Available() bool {
	return e.A.Load()
}

var once sync.Once
var pool *Pool

func InitPool() *Pool {
	once.Do(func() {
		fd, err := os.Open("./pool.txt")
		if err != nil {
			panic(err)
		}
		defer fd.Close()

		data, _ := ioutil.ReadAll(fd)
		pool = &Pool{}

		var wg sync.WaitGroup
		for _, v := range strings.Split(string(data), "\n") {
			wg.Add(1)

			u := v
			go func() {
				defer wg.Done()

				ok, _ := check(u)
				if ok {
					pool.lock.Lock()
					e := Endpoints{U: u}
					e.A.Store(true)
					pool.head = append(pool.head, &e)
					pool.lock.Unlock()
				}
			}()
		}
		wg.Wait()

		fmt.Println("endpoints num is:", len(pool.head))
	})

	return pool
}

func check(e string) (bool, int64) {
	req, err := http.NewRequest("POST", e, bytes.NewReader([]byte(`{"text":"hello",
		"source_lang":"EN",
		"target_lang":"ZH"}`)))
	if err != nil {
		return false, 0
	}
	client := &http.Client{}
	client.Timeout = 10 * time.Second

	now := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return false, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, 0
	}
	data, _ := ioutil.ReadAll(resp.Body)
	if gjson.GetBytes(data, "code").Int() != 200 {
		return false, 0
	}
	if gjson.GetBytes(data, "data").String() == "" {
		return false, 0
	}
	return true, time.Now().Sub(now).Milliseconds()
}

func (p *Pool) current() int64 {
	return p.index.Add(1)
}

func (p *Pool) Pick() *Endpoints {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if len(p.head) == 0 {
		return nil
	}

	return p.head[int(p.current())%len(p.head)]
}

func (p *Pool) Report(e *Endpoints) {
	e.A.Store(false)

	p.lock.Lock()
	for i := range p.head {
		if p.head[i].U == e.U {
			if i == len(p.head)-1 {
				p.head = p.head[:i]
			} else {
				p.head = append(p.head[:i], p.head[i+1:]...)
			}
			break
		}
	}
	p.lock.Unlock()
	fmt.Println("endpoints num is:", len(p.head))

	go func() {
		for range time.Tick(time.Minute) {
			ok, _ := check(e.U)
			if ok {
				e.A.Store(true)
				p.lock.Lock()
				p.head = append(p.head, e)
				p.lock.Unlock()
				break
			}
		}
	}()
}
