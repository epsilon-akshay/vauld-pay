package main

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"vauld-pay/handler"
)

func main() {
	r := mux.NewRouter()
	redisConn := RedisPool()
	service := handler.Service{
		Client: redisConn,
	}

	r.HandleFunc("/kyc", handler.KycPost(service)).Methods(http.MethodPost)
	r.HandleFunc("/kyc", handler.KycGet(service)).Methods(http.MethodGet)
	r.HandleFunc("/deposit", handler.Transfer(service)).Methods(http.MethodPost)

	fmt.Println("starting server at port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}

}

func RedisPool() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return rdb
}
