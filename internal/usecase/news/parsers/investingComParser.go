package parsers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
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

//
//func (p *InvestingComParser) ParseNewsDigest(body string) ([]entity.NewsDigest, error) {
//	p.log.Info("investing.com news parsing")
//
//	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
//	if err != nil {
//		p.log.Warn("Ошибка при парсинге HTML", slog.String("error", err.Error()))
//		return nil, err
//	}
//
//	var newsArray []entity.NewsDigest
//	doc.Find("div.news-analysis-v2_content__z0iLP").Each(func(i int, s *goquery.Selection) {
//		title := s.Find("a").Text()
//		link, exists := s.Find("a").Attr("href")
//		source := s.Find("span[data-test='news-provider-name']").Text()
//		datetimeStr := doc.Find("time").Text()
//
//		parsedTime := p.parseTime(datetimeStr)
//
//		if exists {
//			newsArray = append(newsArray, entity.NewsDigest{
//				Title:       title,
//				Link:        link,
//				Source:      source,
//				PublishedAt: parsedTime,
//			})
//		} else {
//			p.log.Warn("news link not found")
//		}
//	})
//	return newsArray, nil
//
//}

func (p *InvestingComParser) ParseNewsDigest(body string) ([]entity.NewsDigest, error) {
	p.log.Info("investing.com news parsing")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		p.log.Warn("Ошибка при парсинге HTML", slog.String("error", err.Error()))
		return nil, err
	}

	var newsArray []entity.NewsDigest
	doc.Find("div[class='block w-full sm:flex-1 ']").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a[data-test='article-title-link']").Text()
		link, exists := s.Find("a[data-test='article-title-link']").Attr("href")

		source := s.Find("span[data-test='news-provider-name']").Text()
		datetimeStr := s.Find("time[class='ml-2']").Text()

		parsedTime := p.parseTime(datetimeStr)

		if exists {
			newsArray = append(newsArray, entity.NewsDigest{
				Title:       title,
				Link:        link,
				Source:      source,
				PublishedAt: parsedTime,
			})
		} else {
			p.log.Warn("news link not found", slog.String("link", link), slog.String("title", title))
		}
	})
	return newsArray, nil

}

func (p *InvestingComParser) ParseNews(body string, newsDigest entity.NewsDigest) (*entity.News, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		p.log.Warn("Ошибка при парсинге HTML", slog.String("error", err.Error()))
		return nil, err
	}

	newsDiv := doc.Find("div.article_container")

	var newsBuilder strings.Builder
	newsDiv.Find("p").Each(func(i int, s *goquery.Selection) {
		paragraphText := s.Text()
		newsBuilder.WriteString(paragraphText)
		newsBuilder.WriteString("\n")
	})
	return &entity.News{
		Title:       newsDigest.Title,
		Link:        newsDigest.Link,
		Source:      newsDigest.Source,
		Text:        newsBuilder.String(),
		PublishedAt: newsDigest.PublishedAt,
	}, nil

}

func (p *InvestingComParser) FetchHTML(url string) (string, error) {

	res, err := http.Get(url)
	if err != nil {
		p.log.Warn("Error fetching HTML:", slog.String("error", err.Error()))
		return "", fmt.Errorf("error fetching HTML: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		p.log.Warn("invalid status code: %d", res.StatusCode)
		return "", fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		p.log.Warn("Error read body", slog.String("error", err.Error()))
		return "", fmt.Errorf("error read body: %v", err)
	}

	return string(body), nil
}

func (p *InvestingComParser) parseTime(timeStr string) time.Time {
	now := time.Now()

	relativeTime := p.parseRelativeTime(timeStr)
	if !relativeTime.IsZero() {
		return relativeTime
	}

	parsedDate, err := p.parseDate(timeStr)
	if err == nil {
		return parsedDate
	}

	return now
}

func (p *InvestingComParser) parseRelativeTime(relativeTime string) time.Time {
	now := time.Now()

	minutesRegex := regexp.MustCompile(`(\d+)\s*минут(ы|у|)\sназад`)
	hoursRegex := regexp.MustCompile(`(\d+)\s*час(а|ов|)\sназад`)
	daysRegex := regexp.MustCompile(`(\d+)\s*д(е)н(я|ей|ь)\sназад`)

	if matches := minutesRegex.FindStringSubmatch(relativeTime); len(matches) > 1 {
		minutes := p.parseNumber(matches[1])
		return now.Add(-time.Duration(minutes) * time.Minute)
	}

	if matches := hoursRegex.FindStringSubmatch(relativeTime); len(matches) > 1 {
		hours := p.parseNumber(matches[1])
		return now.Add(-time.Duration(hours) * time.Hour)
	}

	if matches := daysRegex.FindStringSubmatch(relativeTime); len(matches) > 1 {
		days := p.parseNumber(matches[1])
		fmt.Println(days)
		return now.AddDate(0, 0, -days)
	}

	return time.Time{}
}

func (p *InvestingComParser) parseDate(dateStr string) (time.Time, error) {
	monthMap := map[string]time.Month{
		"янв.":  time.January,
		"февр.": time.February,
		"мар.":  time.March,
		"апр.":  time.April,
		"мая":   time.May,
		"июн.":  time.June,
		"июл.":  time.July,
		"авг.":  time.August,
		"сент.": time.September,
		"окт.":  time.October,
		"нояб.": time.November,
		"дек.":  time.December,
	}

	dayMap := map[time.Month]int{
		time.January:   31,
		time.February:  29,
		time.March:     31,
		time.April:     30,
		time.May:       31,
		time.June:      30,
		time.July:      31,
		time.August:    31,
		time.September: 30,
		time.October:   31,
		time.November:  30,
		time.December:  31,
	}

	parts := strings.Split(strings.Trim(dateStr, " "), " ")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("не удалось распарсить дату: %s", dateStr)
	}

	day := p.parseNumber(parts[0])
	month, ok := monthMap[parts[1]]
	if !ok {
		return time.Time{}, fmt.Errorf("неизвестный месяц: %s", parts[1])
	}
	if day > dayMap[month] || day <= 0 {
		return time.Time{}, fmt.Errorf("некорректный день: %d", day)
	}

	year := p.parseNumber(parts[2])

	return time.Date(year, month, day, 0, 0, 0, 0, time.Local), nil
}

func (p *InvestingComParser) parseNumber(numStr string) int {
	var num int
	fmt.Sscanf(numStr, "%d", &num)
	return num
}
