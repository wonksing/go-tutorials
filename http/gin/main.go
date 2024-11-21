package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wonksing/go-tutorials/http/gin/handler"
	"github.com/wonksing/go-tutorials/http/gin/middleware"
)

/*
기본 허용 헤더
Accept
Accept-Language
Content-Language
Content-Type (단, 값이 application/x-www-form-urlencoded, multipart/form-data, 또는 text/plain인 경우에만)
*/
func main() {
	r := gin.Default()
	r.Use(middleware.Cors())

	r.GET("/ping", middleware.NoCache(), handler.Ping)

	userHandler := handler.NewUserHandler()
	r.GET("/users/:userId", middleware.NoCache(), userHandler.GetUser)

	server := &http.Server{
		Addr:         ":8080",
		TLSConfig:    &tls.Config{},
		WriteTimeout: time.Duration(10) * time.Second,
		ReadTimeout:  time.Duration(10) * time.Second,
		Handler:      r,
	}

	handleSignals(context.Background(), server, syscall.SIGINT, syscall.SIGTERM)
	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}

}

func handleSignals(ctx context.Context, httpServer *http.Server, signals ...os.Signal) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, signals...)
	go func() {
		<-sigs
		httpServer.Shutdown(ctx)
	}()
}
