package microbot

import (
	"container/list"
	"net/http"
	"sync"
	"time"

	"github.com/pangpanglabs/microbot/common"
)

const DefaultListMax = 1000

var (
	lock         sync.RWMutex
	keyEventList KeyEventList
)

type KeyEvent struct {
	Type    string
	Content string
	Time    time.Time
}

type KeyEventList struct {
	data *list.List
	max  int
}

func KeyEventController() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v []interface{}
		for iter := keyEventList.data.Back(); iter != nil; iter = iter.Prev() {
			v = append(v, iter.Value)
		}
		common.RenderDataJson(w, http.StatusOK, v)
	})
}

func (KeyEventList) SetLength(max int) {
	if max < 1 {
		panic("microbot: invalid queue length")
	}
	keyEventList.max = max
}

func (KeyEventList) New(t string, c string) {
	go keyEventList.push(KeyEvent{
		Type:    t,
		Content: c,
		Time:    time.Now(),
	})
}

func (q *KeyEventList) push(v interface{}) {
	if q.data.Len() >= q.max {
		q.pop()
	}
	defer lock.Unlock()
	lock.Lock()
	q.data.PushFront(v)
}

func (q *KeyEventList) pop() interface{} {
	defer lock.Unlock()
	lock.Lock()
	iter := q.data.Back()
	v := iter.Value
	q.data.Remove(iter)
	return v
}
