package interfaces

import "AIChallenge/internal/entity"

type NewsRepository interface {
	AddNews(news entity.News) error
	GetNewsById(id int) (*entity.News, error)
	ContainNews(url string) (bool, error)
	GetNewsByUrl(url string) (*entity.News, error)
	GetLatestNews(k int) ([]entity.News, error)
}

type Parser interface {
	Parse(body string) ([]entity.News, error)
}

type NewsUseCase interface {
	Start()
	Stop()
	GetLatestNews(numberNews int) ([]entity.News, error)
}
