package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
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
	//r.HandleFunc("/deposit", handler.KycPost(service)).Methods(http.MethodPost)

	fmt.Println("starting server at port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}

}

func RedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   50,
		MaxActive: 10000,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", ":6379")

			// Connection error handling
			if err != nil {
				panic(err)
			}
			return conn, err
		},
	}
}
