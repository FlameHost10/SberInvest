package usecase

import (
	"AIChallenge/internal/entity"
	"AIChallenge/internal/interfaces"
	"AIChallenge/internal/storage"
	"AIChallenge/internal/usecase/news/parsers"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type NewsUseCase struct {
	log          *slog.Logger
	repo         *storage.NewsRepository
	numberNews   int
	interval     time.Duration
	urls         []string
	parsersArray map[string]interfaces.Parser
	stopChan     chan struct{}
}

func NewNewsUseCase(log *slog.Logger, repo *storage.NewsRepository, numberNews int, interval time.Duration) (*NewsUseCase, error) {
	var urls = []string{
		"https://ru.investing.com/news/stock-market-news/",
		"https://ru.tradingview.com/news/",
	}
	var parsersArray = map[string]interfaces.Parser{
		"ru.investing.com":   parsers.NewInvestingComParser(log),
		"ru.tradingview.com": parsers.NewTradingviewComParser(log),
	}

	return &NewsUseCase{
		log:          log,
		repo:         repo,
		numberNews:   numberNews,
		interval:     interval,
		urls:         urls,
		parsersArray: parsersArray,
	}, nil
}

func (uc *NewsUseCase) Start() {
	ticker := time.NewTicker(uc.interval)
	for {
		select {
		case <-ticker.C:
			err := uc.scrapeAndStoreNews()
			if err != nil {
				uc.log.Warn("failed to scrape and store new news data", slog.String("error", err.Error()))
				uc.Stop()
			}
			break

		case <-uc.stopChan:
			return
		}
	}
}

func (uc *NewsUseCase) Stop() {
	uc.log.Info("stopping news useCase")
	uc.stopChan <- struct{}{}
}

func (uc *NewsUseCase) GetLatestNews(numberNews int) ([]entity.News, error) {
	news, err := uc.repo.GetLatestNews(numberNews)
	if err != nil {
		uc.log.Warn("failed to get latest news", slog.String("error", err.Error()))
		return nil, err
	}
	return news, nil
}

func (uc *NewsUseCase) scrapeAndStoreNews() error {
	newsList, err := uc.getNewsFromSites()
	if err != nil {
		uc.log.Warn("Error scraping sites", slog.String("error", err.Error()))
		return err
	}

	for _, newsItem := range newsList {
		exists, err := uc.repo.ContainNews(newsItem.Link)
		if err != nil {
			uc.log.Warn("Error checking news existence", slog.String("error", err.Error()))
			return err
		}

		if !exists {
			err = uc.repo.AddNews(newsItem)
			if err != nil {
				uc.log.Warn("Error adding news", slog.String("error", err.Error()))
				return err
			}
		}
	}
	return nil
}

func (uc *NewsUseCase) fetchHTML(url string) (string, error) {

	res, err := http.Get(url)
	if err != nil {
		uc.log.Warn("Error fetching HTML:", slog.String("error", err.Error()))
		return "", fmt.Errorf("error fetching HTML: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		uc.log.Warn("invalid status code: %d", res.StatusCode)
		return "", fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		uc.log.Warn("Error read body", slog.String("error", err.Error()))
		return "", fmt.Errorf("error read body: %v", err)
	}

	return string(body), nil
}

func (uc *NewsUseCase) getNewsFromSites() ([]entity.News, error) {
	var wg sync.WaitGroup
	newsChannel := make(chan []entity.News)

	for _, url := range uc.urls {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			domain := uc.getDomainFromURL(url)

			parser, exists := uc.parsersArray[domain]
			if !exists {
				uc.log.Warn("Парсер для домена не найден", slog.String("domain", domain))
				return
			}

			html, err := uc.fetchHTML(url)
			if err != nil {
				uc.log.Warn("Ошибка получения HTML", slog.String("error", err.Error()),
					slog.String("url", url))
				return
			}

			news, err := parser.Parse(html)
			if err != nil {
				uc.log.Warn("Ошибка парсинга для домена", slog.String("error", err.Error()),
					slog.String("domain", domain))
				return
			}

			newsChannel <- news
		}(url)
	}

	go func() {
		wg.Wait()
		close(newsChannel)
	}()

	newsArray := make([]entity.News, 0)
	for newsBatch := range newsChannel {
		for _, news := range newsBatch {
			newsArray = append(newsArray, news)
		}
	}
	return newsArray, nil
}

func (uc *NewsUseCase) getDomainFromURL(userUrl string) string {
	u, err := url.Parse(userUrl)
	if err != nil {
		log.Printf("Ошибка при парсинге URL: %v", err)
		return ""
	}
	return u.Host
}
