package usecase

import (
	"AIChallengeNewsAPI/internal/entity"
	"AIChallengeNewsAPI/internal/interfaces"
	"AIChallengeNewsAPI/internal/repository"
	"AIChallengeNewsAPI/internal/usecase/news/parsers"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"sync"
	"time"
)

type NewsUseCase struct {
	log          *slog.Logger
	repo         *repository.Repository
	numberNews   int
	interval     time.Duration
	urls         []string
	parsersArray map[string]interfaces.Parser
	stopChan     chan struct{}
}

func NewNewsUseCase(log *slog.Logger, repo *repository.Repository, numberNews int, interval time.Duration) (*NewsUseCase, error) {
	var urls = []string{
		"https://ru.investing.com/news/",
		"https://www.finmarket.ru/news/",
	}
	var parsersArray = map[string]interfaces.Parser{
		"ru.investing.com": parsers.NewInvestingComParser(log),
		"www.finmarket.ru": parsers.NewFinmarketComParser(log),
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

func (ucNews *NewsUseCase) Start() {
	ticker := time.NewTicker(ucNews.interval)

	err := ucNews.scrapeAndStoreNews()
	if err != nil {
		ucNews.log.Warn("failed to scrape and store new news data", slog.String("error", err.Error()))
		ucNews.Stop()
	}

	for {
		select {
		case <-ticker.C:
			err := ucNews.scrapeAndStoreNews()
			if err != nil {
				ucNews.log.Warn("failed to scrape and store new news data", slog.String("error", err.Error()))
				ucNews.Stop()
			}
			break

		case <-ucNews.stopChan:
			return
		}
	}
}

func (ucNews *NewsUseCase) Stop() {
	ucNews.log.Info("stopping news useCase")
	ucNews.stopChan <- struct{}{}
}

func (ucNews *NewsUseCase) GetLatestNews(numberNews int) ([]entity.News, error) {
	news, err := ucNews.repo.GetLatestNews(numberNews)
	if err != nil {
		ucNews.log.Warn("failed to get latest news", slog.String("error", err.Error()))
		return nil, err
	}
	return news, nil
}

func (ucNews *NewsUseCase) scrapeAndStoreNews() error {
	newsList, err := ucNews.getNewsFromSites()
	if err != nil {
		ucNews.log.Warn("Error scraping sites", slog.String("error", err.Error()))
		return err
	}

	for _, newsItem := range newsList {
		exists, err := ucNews.repo.ContainNews(newsItem.Link)
		if err != nil {
			ucNews.log.Warn("Error checking news existence", slog.String("error", err.Error()))
			return err
		}

		if !exists {
			err = ucNews.repo.AddNews(newsItem)
			if err != nil {
				ucNews.log.Warn("Error adding news", slog.String("error", err.Error()))
				return err
			}
		}
	}
	ucNews.log.Info("The parsing is over")
	return nil
}

func (ucNews *NewsUseCase) getNewsFromSites() ([]entity.News, error) {
	var wg sync.WaitGroup
	newsChannel := make(chan []entity.News)

	for _, url := range ucNews.urls {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			domain := ucNews.getDomainFromURL(url)

			parser, exists := ucNews.parsersArray[domain]
			if !exists {
				ucNews.log.Warn("Парсер для домена не найден", slog.String("domain", domain))
				return
			}

			newsDigests, err := ucNews.getNewsDigestFromSite(url, parser)
			if err != nil {
				ucNews.log.Warn("Ошибка получения digests", slog.String("error", err.Error()))
				return
			}

			news, err := ucNews.getNewsFromNewsDigest(newsDigests, parser)
			if err != nil {
				ucNews.log.Warn("Ошибка получения news", slog.String("error", err.Error()))
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

func (ucNews *NewsUseCase) getNewsDigestFromSite(url string, parser interfaces.Parser) ([]entity.NewsDigest, error) {
	html, err := parser.FetchHTML(url)
	if err != nil {
		ucNews.log.Warn("Ошибка получения HTML", slog.String("error", err.Error()),
			slog.String("url", url))
		return nil, fmt.Errorf("error fetching HTML: %v", err)
	}

	news, err := parser.ParseNewsDigest(html)
	if err != nil {
		ucNews.log.Warn("Ошибка парсинга для урла", slog.String("error", err.Error()),
			slog.String("url", url))
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}
	return news, nil
}

func (ucNews *NewsUseCase) getNewsFromNewsDigest(newsDigest []entity.NewsDigest, parser interfaces.Parser) ([]entity.News, error) {
	var newsArray []entity.News
	for _, newsItem := range newsDigest {
		html, err := parser.FetchHTML(newsItem.Link)
		if err != nil {
			ucNews.log.Warn("Ошибка получения HTML", slog.String("error", err.Error()),
				slog.String("url", newsItem.Link))
			return nil, fmt.Errorf("error fetching HTML: %v", err)
		}
		news, err := parser.ParseNews(html, newsItem)
		if err != nil {
			ucNews.log.Warn("Ошибка парсинга новостей", slog.String("error", err.Error()))
			return nil, fmt.Errorf("error parsing HTML: %v", err)
		}

		newsArray = append(newsArray, *news)

	}
	return newsArray, nil
}

func (ucNews *NewsUseCase) getDomainFromURL(userUrl string) string {
	u, err := url.Parse(userUrl)
	if err != nil {
		log.Printf("Ошибка при парсинге URL: %v", err)
		return ""
	}
	return u.Host
}
