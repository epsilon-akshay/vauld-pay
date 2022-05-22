package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"math/rand"
	"net/http"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type Body struct {
	EmailID string `json:"emailID"`
	Name    string `json:"name"`
	IdProof string `json:"id"`
}

type Value struct {
	WalletAdd []string `json:"walletAddress"`
	Name      string   `json:"name"`
	Id        string   `json:"id"`
}

type Service struct {
	Client *redis.Pool
}

func KycPost(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rand.Seed(time.Now().UnixNano())

		decoder := json.NewDecoder(r.Body)
		var t Body
		err := decoder.Decode(&t)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s := fmt.Sprintf("{success:false, err:%s}", err.Error())
			w.Write([]byte(s))
			return
		}
		walletAddress := RandStringRunes(10)
		v := Value{
			WalletAdd: []string{walletAddress},
			Name:      t.Name,
			Id:        t.IdProof,
		}

		_, err = s.Client.Get().Do("SET", t.EmailID, v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s := fmt.Sprintf("{success:false, err:%s}", err.Error())
			w.Write([]byte(s))
			return
		}

		w.WriteHeader(http.StatusOK)
		l := fmt.Sprintf("{success:true, data:%s}", walletAddress)
		w.Write([]byte(l))
		return
	}
}

func KycGet(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query()["emailID"][0]
		_, err := redis.String(s.Client.Get().Do("GET", key))
		if err == redis.ErrNil {
			w.WriteHeader(http.StatusInternalServerError)
			s := fmt.Sprintf("{success:false, err:%s}", err.Error())
			w.Write([]byte(s))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s := fmt.Sprintf("{success:false, err:%s}", err.Error())
			w.Write([]byte(s))
			return
		}

		w.WriteHeader(http.StatusOK)
		l := fmt.Sprintf("{success:true}")
		w.Write([]byte(l))
		return
	}
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
