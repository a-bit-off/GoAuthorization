package main

import (
	"app/internal/application"
	"app/internal/repository"
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()
	dbpool, err := repository.InitDBConn(ctx)
	if err != nil {
		log.Fatalf("%w failed to init DB connection", err)
	}
	defer dbpool.Close()

	a := application.NewApp(ctx, dbpool)

	// создаем роутер
	r := httprouter.New()
	a.Routes(r)
	srv := &http.Server{Addr: "0.0.0.0:8080", Handler: r}
	fmt.Println("It is alive!")
	srv.ListenAndServe()
}
