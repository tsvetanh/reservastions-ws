package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reservations/configuration"
	"reservations/services/hall"
	"reservations/services/user"
	"syscall"
	"time"
)

func main() {
	d, err := configuration.Init()
	if err != nil {
		panic(err)
	}

	r := Routes(d)

	srv := &http.Server{
		Addr:         ":" + d.Cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  240 * time.Second,
	}

	go func() {
		var err error
		err = srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	go configuration.KeepConnectionsAlive(d.Db, time.Minute*5)

	err = d.Db.AutoMigrate(user.User{}, user.UserRoles{}, user.Role{}, hall.HallImage{})
	if err != nil {
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")

}
