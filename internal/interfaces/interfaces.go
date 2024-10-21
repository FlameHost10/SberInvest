package interfaces

import (
	"AIChallenge/internal/entity"
)

type RepositoryInter interface {
	AddNews(news entity.News) error
	GetNewsById(id int) (*entity.News, error)
	ContainNews(url string) (bool, error)
	GetNewsByUrl(url string) (*entity.News, error)
	GetLatestNews(k int) ([]entity.News, error)
}

type Parser interface {
	ParseNewsDigest(body string) ([]entity.NewsDigest, error)
	ParseNews(body string, newsDigest entity.NewsDigest) (*entity.News, error)
	FetchHTML(url string) (string, error)
}

type NewsUseCase interface {
	Start()
	Stop()
	GetLatestNews(numberNews int) ([]entity.News, error)
}
