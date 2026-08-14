package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/pablo/backend/internal/api"
	"github.com/pablo/backend/internal/auth"
	"github.com/pablo/backend/internal/config"
	"github.com/pablo/backend/internal/integration"
	"github.com/pablo/backend/internal/service"
	"github.com/pablo/backend/internal/storage"
	"github.com/pablo/backend/schema"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	defer func() {
		if err := recover(); err != nil {
			log.Println("PANIC error occurred:", err)
		}
	}()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal("error while init config:", err.Error())
	}

	log.Println("DEBUG:", cfg.DebugFlag.Flag)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("failed to connect to database:", err.Error())
	}

	if err = db.Ping(); err != nil {
		log.Fatal("failed to ping database:", err.Error())
	}
	log.Println("connected to db")

	dbx := sqlx.NewDb(db, "postgres")

	storageRegistry := storage.NewStorageRegistry(dbx)

	integrationRegistry, err := integration.NewIntegrationRegistry(cfg)
	if err != nil {
		log.Fatal("failed to init integration registry:", err.Error())
	}

	serviceRegistry := service.NewServiceRegistry(cfg, storageRegistry, integrationRegistry)

	authManager := auth.NewAuthManager(cfg.Auth)

	httpServer := api.NewHttpServer(serviceRegistry, authManager)

	StartHttpServer(ctx, authManager, httpServer)
}

func StartHttpServer(
	ctx context.Context,
	authManager auth.AuthManager,
	server schema.ServerInterface,
) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("PANIC error occurred: ", err)
		}
	}()

	handler := schema.HandlerWithOptions(server, schema.ChiServerOptions{
		BaseURL:     "",
		Middlewares: nil,
	})

	authMiddleware := authManager.AuthMiddleware()

	router := chi.NewRouter()
	router.Use(corsMiddleware)

	router.Group(func(router chi.Router) {
		router.Handle("/v1/login", handler)
	})

	router.Group(func(router chi.Router) {
		router.Use(authMiddleware)

		router.Handle("/v1/posts", handler)
		router.Handle("/v1/posts/{post_id}", handler)
		router.Handle("/v1/posts/{post_id}/publish", handler)
		router.Handle("/v1/posts/{post_id}/stats", handler)
		router.Handle("/v1/posts/{post_id}/media", handler)
		router.Handle("/v1/accounts", handler)
		router.Handle("/v1/accounts/sync", handler)
		router.Handle("/v1/accounts/connect_url", handler)
	})

	group, ctx := errgroup.WithContext(ctx)

	httpServer := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	group.Go(func() error {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	group.Go(func() error {
		<-ctx.Done()
		return httpServer.Shutdown(context.Background())
	})

	if err := group.Wait(); err != nil {
		log.Fatal("err while init http server:", err.Error())
	}
}

// corsMiddleware пускает фронт с другого origin — в дев-режиме vite крутится
// на своём порту.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
