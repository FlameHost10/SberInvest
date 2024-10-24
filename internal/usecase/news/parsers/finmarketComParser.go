package parsers

import (
	"AIChallengeNewsAPI/internal/entity"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type FinmarketComParser struct {
	log *slog.Logger
}

func NewFinmarketComParser(log *slog.Logger) *FinmarketComParser {
	return &FinmarketComParser{log: log}
}

func (p *FinmarketComParser) ParseNewsDigest(body string) ([]entity.NewsDigest, error) {
	p.log.Info("finmarket.com news parsing")

	// Парсим документ с учетом кодировки
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении HTML-документа: %v", err)
	}

	var newsList []entity.NewsDigest

	// Ищем все блоки с новостями в div.ind_article
	doc.Find("div.ind_article").Each(func(i int, s *goquery.Selection) {
		// Переменные для хранения текущей даты и заголовка
		var currentDate time.Time
		var news entity.NewsDigest

		// Проходим по дочерним элементам внутри div.ind_article
		s.Children().Each(func(i int, selection *goquery.Selection) {
			if selection.Is("span.date") {
				// Если это дата, парсим её
				dateStr := selection.Text()
				date, err := p.parseDate(dateStr)
				if err != nil {
					log.Printf("Ошибка парсинга даты: %v", err)
					return
				}
				currentDate = date
			} else if selection.Is("div.title") {
				selection.Find("a").Each(func(i int, a *goquery.Selection) {
					title := a.Text()
					link, exists := a.Attr("href")
					if !exists {
						log.Printf("Не удалось найти ссылку для новости")
						return
					}
					fullLink := fmt.Sprintf("https://finmarket.ru%s", link)

					news = entity.NewsDigest{
						Title:       title,
						Link:        fullLink,
						PublishedAt: currentDate,
						Source:      "finmarket.ru",
					}
					newsList = append(newsList, news)
				})
			}
		})
	})

	return newsList, nil
}

func (p *FinmarketComParser) parseDate(dateStr string) (time.Time, error) {
	replacer := strings.NewReplacer(
		"января", "January",
		"февраля", "February",
		"марта", "March",
		"апреля", "April",
		"мая", "May",
		"июня", "June",
		"июля", "July",
		"августа", "August",
		"сентября", "September",
		"октября", "October",
		"ноября", "November",
		"декабря", "December",
	)

	dateStr = replacer.Replace(dateStr)

	dateStr = strings.TrimSpace(dateStr)

	layout := "2 January 2006 года 15:04"
	parsedDate, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDate, nil
}

func (p *FinmarketComParser) ParseNews(body string, newsDigest entity.NewsDigest) (*entity.News, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		p.log.Warn("Ошибка при парсинге HTML", slog.String("error", err.Error()))
		return nil, err
	}

	var contentBuilder strings.Builder

	doc.Find("div.body").Each(func(i int, s *goquery.Selection) {
		content := s.Text()
		contentBuilder.WriteString(content)
		contentBuilder.WriteString("\n")
	})

	return &entity.News{
		Title:       newsDigest.Title,
		Link:        newsDigest.Link,
		Source:      newsDigest.Source,
		Text:        contentBuilder.String(),
		PublishedAt: newsDigest.PublishedAt,
	}, nil
}

func (p *FinmarketComParser) FetchHTML(url string) (string, error) {

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

	utf8Reader := transform.NewReader(res.Body, charmap.Windows1251.NewDecoder()) // Читаем декодированное тело

	body, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		p.log.Warn("Error read body", slog.String("error", err.Error()))
		return "", fmt.Errorf("error read body: %v", err)
	}

	return string(body), nil
}
