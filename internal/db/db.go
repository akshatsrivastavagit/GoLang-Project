package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	dbPool      *pgxpool.Pool
	redisClient *redis.Client
)

type DB interface {
	Exec(ctx context.Context, sql string, args ...interface{}) error
	Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
}

type Redis interface {
	Publish(ctx context.Context, channel string, message interface{}) error
	XAdd(ctx context.Context, args *redis.XAddArgs) *redis.StringCmd
	XRead(ctx context.Context, args *redis.XReadArgs) *redis.XStreamSliceCmd
}

type DBWrapper struct {
	pool *pgxpool.Pool
}

func (w *DBWrapper) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := w.pool.Exec(ctx, sql, args...)
	return err
}

func (w *DBWrapper) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	return w.pool.Query(ctx, sql, args...)
}

type RedisWrapper struct {
	client *redis.Client
}

func (w *RedisWrapper) Publish(ctx context.Context, channel string, message interface{}) error {
	// Convert message to JSON before publishing
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	cmd := w.client.Publish(ctx, channel, data)
	return cmd.Err()
}

func (w *RedisWrapper) XAdd(ctx context.Context, args *redis.XAddArgs) *redis.StringCmd {
	// Convert values to JSON strings
	values := make(map[string]interface{})
	if args.Values != nil {
		if m, ok := args.Values.(map[string]interface{}); ok {
			for k, v := range m {
				data, err := json.Marshal(v)
				if err != nil {
					log.Printf("Error marshaling value for key %s: %v", k, err)
					continue
				}
				values[k] = string(data)
			}
		}
	}
	args.Values = values
	return w.client.XAdd(ctx, args)
}

func (w *RedisWrapper) XRead(ctx context.Context, args *redis.XReadArgs) *redis.XStreamSliceCmd {
	return w.client.XRead(ctx, args)
}

func InitDB() error {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("unable to parse database config: %v", err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}

	dbPool = pool
	log.Println("Successfully connected to database")
	return nil
}

func InitRedis() error {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPass := os.Getenv("REDIS_PASSWORD")
	redisDB := 0

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       redisDB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("unable to connect to Redis: %v", err)
	}

	redisClient = client
	log.Println("Successfully connected to Redis")
	return nil
}

func CloseDB() {
	if dbPool != nil {
		dbPool.Close()
	}
}

func CloseRedis() {
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}
}

func GetDB() DB {
	return &DBWrapper{pool: dbPool}
}

func GetRedis() Redis {
	return &RedisWrapper{client: redisClient}
}

type Rows interface {
	Close()
	Next() bool
	Scan(dest ...interface{}) error
} 