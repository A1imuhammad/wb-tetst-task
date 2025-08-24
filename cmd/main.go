package main

import (
	"demoserv/internal/cache"
	"demoserv/internal/config"
	"demoserv/internal/http-server/handlers/getOrder"
	"demoserv/internal/kafka"
	"demoserv/internal/postgress"
	
	"net/http"
	"context"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()
	// читаем конфиг
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	log.Println("config read")

	// подключаемся к базе
	pool, err := postgress.New(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	log.Println("database connected")
	defer pool.Close()

	// инициализируем кэш
	ordersCache := cache.NewCache(1000)
	if err := cache.InitCacheFromDB(ctx, pool, ordersCache); err != nil {
		log.Fatalf("unable to init cache from database: %v", err)
	}
	log.Println("cache initialized")

	// инициализируем kafka consumer
	log.Println("Starting Kafka consumer in background...")
	go func() {
		kafka.NewConsumer(ctx, cfg, pool, ordersCache)
	}()

	go func() {
		kafka.NewProducer(cfg)
	}()




	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"*", "null"}, 
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: false,
        MaxAge:           300, 
    }))
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	

	router.Get("/order/{order_uid}", getorder.New(ctx, ordersCache, pool))

	log.Printf("starting server on %s", cfg.HttpServer.Address)

	srv := &http.Server{
		Addr:        cfg.HttpServer.Address,
		Handler:     router,
		ReadTimeout: cfg.HttpServer.Timeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %s", err.Error())
	}

}
