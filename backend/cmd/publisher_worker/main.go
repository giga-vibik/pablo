// Воркер отложенной публикации: раз в тик забирает посты, у которых наступил
// scheduled_at, и публикует их через zernio.
//
// Посты забираются через UPDATE ... FOR UPDATE SKIP LOCKED — статус меняется на
// publishing прямо в выборке, поэтому второй инстанс воркера тот же пост не
// подхватит.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/pablo/backend/internal/config"
	"github.com/pablo/backend/internal/integration"
	"github.com/pablo/backend/internal/service"
	"github.com/pablo/backend/internal/storage"
)

const (
	defaultTickSeconds = 30
	defaultBatchSize   = 10
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

	tick := cfg.Publisher.TickSeconds
	if tick <= 0 {
		tick = defaultTickSeconds
	}

	batchSize := cfg.Publisher.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	log.Printf("publisher worker started: tick=%ds batch=%d", tick, batchSize)

	ticker := time.NewTicker(time.Duration(tick) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("publisher worker stopped")
			return
		case <-ticker.C:
			publishDuePosts(ctx, serviceRegistry, batchSize)
		}
	}
}

func publishDuePosts(ctx context.Context, serviceRegistry *service.Services, batchSize int) {
	posts, err := serviceRegistry.PostService.ListDuePosts(ctx, batchSize)
	if err != nil {
		log.Println("error: while listing due posts", err.Error())
		return
	}

	for i := range posts {
		postID := posts[i].GetID()

		published, pErr := serviceRegistry.PostService.PublishPost(ctx, postID)
		if pErr != nil {
			log.Printf("error: while publishing post %s: %s", postID.String(), pErr.Error())

			if fErr := serviceRegistry.PostService.FailPost(ctx, postID); fErr != nil {
				log.Printf("error: while marking post %s failed: %s", postID.String(), fErr.Error())
			}

			continue
		}

		log.Printf("post %s published with status %s", postID.String(), published.GetStatus())
	}
}
