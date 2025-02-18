package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type repository struct {
	db  *sql.DB
	rdb *redis.Client
}

func New(db *sql.DB, rdb *redis.Client) *repository {
	return &repository{db: db, rdb: rdb}
}

func (r *repository) GetLongUrl(ctx context.Context, short_url string) (string, error) {

	long_url, err := r.GetFromCache(ctx, short_url)
	if err == nil {
		return long_url, nil
	}

	query := `
  SELECT long_url 
  FROM url_table 
  WHERE short_url =  $1
`
	err = r.db.QueryRowContext(ctx, query, short_url).Scan(&long_url)
	if err != nil {
		log.Print(err.Error())
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", ErrNotFound
		default:
			return "", err
		}
	}
	go r.AddToCache(ctx, short_url, long_url)
	return long_url, err

}

func (r *repository) GetFromCache(ctx context.Context, short_url string) (string, error) {
	long_url, err := r.rdb.Get(ctx, short_url).Result()
	if err != nil {
		return "", err
	}
	return long_url, nil
}

func (r *repository) AddToCache(ctx context.Context, short_url string, long_url string) {
	_, err := r.rdb.Set(ctx, short_url, long_url, 24*60*60*time.Second).Result()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("added to cache")
	}
}
