package entity

import "time"

type News struct {
	Title       string
	Link        string
	Source      string
	Text        string
	PublishedAt time.Time
}

type NewsDigest struct {
	Title       string
	Link        string
	Source      string
	PublishedAt time.Time
}

func (n *News) convertToNewsDigest() *NewsDigest {
	return &NewsDigest{
		Title:       n.Title,
		Link:        n.Link,
		Source:      n.Source,
		PublishedAt: n.PublishedAt,
	}
}

type NewsRepo struct {
	ID          int
	Title       string
	Link        string
	Source      string
	PublishedAt time.Time
}
