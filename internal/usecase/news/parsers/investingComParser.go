package parsers

import (
	"log/slog"
	"strings"
	"time"

	"AIChallenge/internal/entity"
	"github.com/PuerkitoBio/goquery"
)

type InvestingComParser struct {
	log *slog.Logger
}

func NewInvestingComParser(log *slog.Logger) *InvestingComParser {
	return &InvestingComParser{log: log}
}

func (p *InvestingComParser) Parse(body string) ([]entity.News, error) {
	p.log.Info("investing.com news")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		p.log.Warn("Ошибка при парсинге HTML", slog.String("error", err.Error()))
		return nil, err
	}

	var newsArray []entity.News
	doc.Find("div.news-analysis-v2_content__z0iLP").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a").Text()
		link, exists := s.Find("a").Attr("href")
		source := s.Find("span[data-test='news-provider-name']").Text()
		datetimeStr := doc.Find("time").AttrOr("datetime", "")
		if datetimeStr == "" {
			p.log.Warn("datetime attribute not found")
		}

		const layout = "2006-01-02 15:04:05"

		parsedTime, err := time.Parse(layout, datetimeStr)
		if err != nil {
			p.log.Warn("failed to parse datetime", slog.String("error", err.Error()))
		}

		if exists {
			newsArray = append(newsArray, entity.News{
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
