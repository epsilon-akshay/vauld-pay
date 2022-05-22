package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"net/http"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type Body struct {
	EmailID   string        `json:"emailID"`
	Name      string        `json:"name"`
	IdProof   string        `json:"id"`
	AccountNs []WalletValue `json:"wallets"`
}

type Value struct {
	WalletAdd []WalletValue `json:"walletAddress"`
	Name      string        `json:"name"`
	Id        string        `json:"id"`
}

// UnmarshalBinary -
func (e Value) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	return nil
}

func (i Value) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

type WalletValue struct {
	AccNo               string   `json:"AccNo"`
	Name                string   `json:"Name"`
	FiatAmountDetails   []Amount `json:"fiatAmount"`
	CryptoAmountDetails []Amount `json:"cryptoAmount"`
	Username            string   `json:"username"`
}

type Amount struct {
	Val      float64 `json:"amount"`
	Currency string  `json:"currency"`
}
type Service struct {
	Client *redis.Client
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
		var walValue []WalletValue
		for _, i := range t.AccountNs {
			username := i.Name + "@vauld.com"
			walValue = append(walValue, WalletValue{AccNo: i.AccNo, CryptoAmountDetails: i.CryptoAmountDetails, FiatAmountDetails: i.FiatAmountDetails, Name: i.Name, Username: username})
		}

		v := Value{
			WalletAdd: walValue,
			Name:      t.Name,
			Id:        t.IdProof,
		}

		redisVal, _ := json.Marshal(v)

		err = s.Client.Set(context.Background(), t.EmailID, redisVal, 0).Err()
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
		val, err := s.Client.Get(context.Background(), key).Result()
		if err == redis.Nil {
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
		fmt.Println(val)
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
