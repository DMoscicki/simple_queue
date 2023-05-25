package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type store struct {
	datastore map[string][]string
	cha chan struct{}
	locker sync.RWMutex
}

type userRequests struct {
	queue *store
}

type Answer struct {
	Value string
	Status string
}


func (ur *userRequests) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ur.Get(w, r)
	case http.MethodPut:
		ur.Put(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Данный метод отсутствует на сервере"))
	}
}

func (user *userRequests) Get(w http.ResponseWriter, r *http.Request) {

	log.Println(r.Method)

	var res = make(chan string, 1)

	key := r.URL.Path[1:]

	if len(key) == 0 {
		w.WriteHeader(404)
	}

	timer := r.URL.Query().Get("timeout")

	sec, err := strconv.Atoi(timer)
	if err != nil {
		sec = 0
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second * time.Duration(sec))
	defer cancel()

	select {
	case <- user.queue.cha:
		user.queueelem(key, res)
		value := <- res
		stat := fmt.Sprintf("статус %v (ok)", http.StatusOK)
		b := Answer{Value: value, 
			Status: stat}
		data, err := json.Marshal(b)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(err.Error()))
		} else {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
	case <- ctx.Done():
		w.WriteHeader(http.StatusBadRequest)
	case <- r.Context().Done():
		w.WriteHeader(http.StatusBadRequest)
	}

}

func (user *userRequests) queueelem(key string, res chan string) {
	var val string
	user.queue.locker.Lock()
	defer user.queue.locker.Unlock()
	v, ok := user.queue.datastore[key]
	if ok {
		if len(v) == 1 {
			val = v[0]
			res <- val
			delete(user.queue.datastore, key)
		} else {
			val = v[0]
			res <- val
			user.queue.datastore[key] = v[1:]
		}
	}
}

func (user *userRequests) Put(w http.ResponseWriter, r *http.Request) {

	log.Println(r.Method)

	f := r.URL.Path

	key := f[1:]

	if string(key[len(key)-1]) == "/" {
		key = strings.Trim(key, "/")
	}

	val := r.URL.Query().Get("v")

	user.queue.locker.RLock()
	defer user.queue.locker.RUnlock()
	user.queue.datastore[key] = append(user.queue.datastore[key], val)
	user.queue.cha <- struct{}{}

	if len(user.queue.cha) == 10 {
		user.queue.cha = nil
		user.queue.cha = make(chan struct{}, 10)
		user.queue.cha <- struct{}{}
	}

	stat := fmt.Sprintf("статус %v (ok) ", http.StatusOK)
	ans := Answer{
		Value: "добавлено " + val,
		Status: stat,
	}

	data, err := json.Marshal(ans)
	if err != nil {
		ans := fmt.Sprintf("ошибка на сервере")
		stat := fmt.Sprintf("статус %v (bad gateaway)", http.StatusBadGateway)
		answer := Answer{
			Value: ans,
			Status: stat,
		}
		data, err := json.Marshal(answer)
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusBadGateway)
		w.Write(data)
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func main() {
	var port string
	
	fmt.Print("Введите порт: ")
	fmt.Fscan(os.Stdin, &port)

	if port == "" {
		log.Fatal("Не ввели порт")
	}

	dt := make(map[string][]string, 0)
	st := make(chan struct{}, 10)

	mux := http.NewServeMux()

	mux.Handle("/", &userRequests{
		&store{
			datastore: dt,
			cha: st,
		},
	})

	log.Fatal(http.ListenAndServe(":"+port, mux))
}