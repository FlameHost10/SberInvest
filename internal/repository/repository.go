package repository

import (
	"AIChallengeNewsAPI/internal/entity"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(connectionString string, log *slog.Logger) (*Repository, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Warn("cannot open to database news", slog.String("error", err.Error()))
		return nil, err
	}

	err = waitForDB(db, log)
	if err != nil {
		log.Warn("cannot connect to database news", slog.String("error", err.Error()))
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (repo *Repository) AddNews(news entity.News) error {
	query := "INSERT INTO news (title, url, source, published_at, text) VALUES ($1, $2, $3, $4, $5)"
	_, err := repo.db.Exec(query, news.Title, news.Link, news.Source, news.PublishedAt, news.Text)
	return err
}

func (repo *Repository) GetNewsById(id int) (*entity.News, error) {
	var news entity.News
	query := "SELECT title, url, source, published_at, text FROM news WHERE id = $1"
	row := repo.db.QueryRow(query, id)
	err := row.Scan(&news.Title, &news.Link, &news.Source, &news.PublishedAt, &news.Text)
	if err != nil {
		return nil, err
	}
	return &news, nil
}

func (repo *Repository) GetNewsByUrl(url string) (*entity.News, error) {
	var news entity.News
	query := "SELECT title, url, source, published_at, text FROM news WHERE url = $1"
	row := repo.db.QueryRow(query, url)
	err := row.Scan(&news.Title, &news.Link, &news.Source, &news.PublishedAt, &news.Text)
	if err != nil {
		return nil, err
	}
	return &news, nil
}

func (repo *Repository) ContainNews(url string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM news WHERE url = $1)"
	err := repo.db.QueryRow(query, url).Scan(&exists)
	return exists, err
}

func (repo *Repository) GetLatestNews(k int) ([]entity.News, error) {
	query := `SELECT title, url, source, published_at, text FROM news
		ORDER BY published_at DESC LIMIT $1`
	rows, err := repo.db.Query(query, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newsList []entity.News
	for rows.Next() {
		var news entity.News
		if err := rows.Scan(&news.Title, &news.Link, &news.Source, &news.PublishedAt, &news.Text); err != nil {
			return nil, err
		}
		newsList = append(newsList, news)
	}

	return newsList, nil
}

func waitForDB(db *sql.DB, log *slog.Logger) error {
	for i := 0; i < 10; i++ {
		err := db.Ping()
		if err == nil {
			return nil
		}
		log.Info("Waiting for database connection")
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("failed to connect to the database")
}
