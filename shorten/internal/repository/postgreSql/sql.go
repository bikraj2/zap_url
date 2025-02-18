package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository struct {
	rdb           *redis.Client
	db            *sql.DB
	kgsServiceUrl string
}

func New(db *sql.DB, rdb *redis.Client, url string) *Repository {
	return &Repository{db: db, kgsServiceUrl: url, rdb: rdb}
}

func (r *Repository) CreateShortUrl(ctx context.Context, longURL string) (string, error) {
	var short_url string
	var err error
	alreadyExist := true
	for alreadyExist {
		short_url, err = r.GetCode(ctx)
		if err != nil {
			return "", err
		}
		alreadyExist, err = r.ExistsInBloomFilter(ctx, short_url)
		if err != nil {
			return "", err
		}
	}
	expiry_time := time.Now().Add(24 * 60 * 60 * time.Second)
	query := `
  INSERT INTO  url_table 
  (long_url,short_url,expires_at) 
  VALUES ($1,$2,$3)
  `
	_, err = r.db.ExecContext(ctx, query, longURL, short_url, expiry_time)
	if err != nil {
		return "", err
	}
	_ = r.AddToBloomFilter(ctx, short_url)
	return short_url, nil
}

func (r *Repository) GetCode(ctx context.Context) (string, error) {

	resp, err := http.Get(r.kgsServiceUrl + "")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get key from KGS")
	}

	var result struct {
		Key string `json:"key"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.Key, nil
}
func (r *Repository) ExistsInBloomFilter(ctx context.Context, shortURL string) (bool, error) {
	exists, err := r.rdb.Do(ctx, "BF.EXISTS", "short_urls", shortURL).Int()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}
func (r *Repository) AddToBloomFilter(ctx context.Context, shortURL string) error {
	return r.rdb.Do(ctx, "BF.ADD", "short_urls", shortURL).Err()
}
func (r *Repository) LoadShortURLsIntoBloomFilter(ctx context.Context) error {
	query := `
  SELECT short_url 
  FROM url_table
  `
	rows, err := r.db.Query(query)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil
		}
		return err
	}
	var key string
	for rows.Next() {
		err = rows.Scan(&key)
		if err != nil {
			return err
		}
		if err = r.AddToBloomFilter(ctx, key); err != nil {
			continue
		}
	}
	fmt.Println("Bloom filter initialized with short URLs")
	return nil
}
