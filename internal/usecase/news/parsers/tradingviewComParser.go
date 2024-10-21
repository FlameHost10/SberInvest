package parsers

import (
	"log"
	"log/slog"
	"strings"
	"time"

	"AIChallenge/internal/entity"
	"github.com/PuerkitoBio/goquery"
)

type TradingviewComParser struct {
	log *slog.Logger
}

func NewTradingviewComParser(log *slog.Logger) *TradingviewComParser {
	return &TradingviewComParser{log: log}
}

func (p *TradingviewComParser) ParseNewsDigest(body string) ([]entity.NewsDigest, error) {
	p.log.Info("tradingview.com news parsing")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Printf("Ошибка при парсинге HTML: %v", slog.String("error", err.Error()))
		return nil, err
	}

	prefixLink := "https://ru.tradingview.com"
	var newsArray []entity.NewsDigest

	doc.Find("a.card-DmjQR0Aa").Each(func(i int, s *goquery.Selection) {
		title := s.Find("div.title-DmjQR0Aa").Text()
		link, exists := s.Attr("href")
		source := s.Find("span.provider-TUPxzdRV").Text()
		link = prefixLink + link
		eventTimeStr := doc.Find("relative-time").AttrOr("event-time", "")
		if eventTimeStr == "" {
			p.log.Warn("event-time attribute not found")
		}
		const layout = "Mon, 02 Jan 2006 15:04:05 MST"

		parsedTime, err := time.Parse(layout, eventTimeStr)
		if err != nil {
			p.log.Warn("failed to parse event-time", slog.String("error", err.Error()))
		}

		if exists {
			newsArray = append(newsArray, entity.NewsDigest{
				Title:       title,
				Link:        link,
				Source:      source,
				PublishedAt: parsedTime,
			})
		} else {
			p.log.Warn("news link not found")
		}
	})

	return newsArray, nil

}
