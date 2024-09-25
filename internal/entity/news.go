package entity

import "time"

type News struct {
	Title       string
	Link        string
	Source      string
	PublishedAt time.Time
}

type NewsRepo struct {
	ID          int
	Title       string
	Link        string
	Source      string
	PublishedAt time.Time
}
