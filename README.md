# News Aggregator API

## Описание проекта

Данный проект представляет собой API для получения и хранения новостей, написанный на языке Go (Golang). Новости хранятся в базе данных PostgreSQL. API поддерживает два основных направления работы:
1. **Обновление новостей**: новостные данные автоматически собираются через периодический парсинг внешних сайтов и сохраняются в базу данных.
2. **Обработка запросов пользователей**: через API можно получить последние новости с возможностью указания количества возвращаемых новостей.

### Основные возможности
- **Параллельный парсинг новостей**: каждая внешняя новостная платформа обрабатывается в отдельной горутине (goroutine), что позволяет обрабатывать несколько источников одновременно без блокировки основного потока.
- **Поддерживаются сайты**:
    - [Investing.com](https://ru.investing.com/news)
    - [Finmarket.ru](https://www.finmarket.ru/news)
- **Настраиваемый интервал обновления**: процесс парсинга новостей происходит каждые `n` минут (значение `n` можно настраивать).
- **Добавление новых новостей в базу данных**: сохраняются только уникальные новости, которых ещё нет в базе.
- **Логирование**: все этапы работы программы логируются, включая ошибки, начало и конец каждого парсинга.
- **Масштабируемость**: возможность легко добавлять новые источники новостей с минимальными изменениями в коде.

### Пример запроса и ответа

Пример запроса для получения двух последних новостей:

```
GET /news?limit=2
```


Ответ (JSON):
```json
[
  {
    "Title": "Рынок акций Норвегии закрылся ростом, Oslo OBX прибавил 0,70%",
    "Link": "https://ru.investing.com/news/stock-market-news/article-2544278",
    "Source": "Investing.com",
    "Text": "Investing.com – Фондовый рынок Норвегии завершил торги четверга ростом...",
    "PublishedAt": "2024-10-24T18:02:52.82962Z"
  },
  {
    "Title": "Средний курс юаня со сроком расчетов 'сегодня' по итогам торгов составил 13,532 руб.",
    "Link": "https://finmarket.ru/currency/news/6274600",
    "Source": "Finmarket.ru",
    "Text": "24 октября. FINMARKET.RU - Средневзвешенный курс юаня со сроком расчетов 'сегодня'...",
    "PublishedAt": "2024-10-24T18:00:00Z"
  }
]
```

Если параметр limit не указан, по умолчанию возвращаются 10 последних новостей.


### Установка и Запуск
**1. Поднятие базы данных**: нДля запуска базы данных PostgreSQL используйте Docker и команду docker-compose:

```bash
docker-compose up --build
```

**2. Запуск API**: После запуска базы данных выполните команду для запуска API:

```bash
go run ./cmd/main/main.go
```
Убедитесь, что вы находитесь в корневой папке проекта при выполнении этой команды.



### Настройка
- Интервал парсинга новостей можно настраивать (по умолчанию парсинг происходит каждые n минут).

- Установка своих config, для использования программы (изменение данных в файле .env)

- Добавление новых сайтов для парсинга: достаточно реализовать новую функцию для обработки HTML-страницы новостного сайта и интегрировать её в проект.


### Логирование
Программа ведет логирование следующих событий:

- Ошибки, возникающие в процессе работы (например, ошибки парсинга, проблемы с базой данных).

-  Начало и конец каждого процесса парсинга новостей с конкретного сайта.


### Масштабируемость
- ***Добавление новых парсеров***
  Проект изначально построен с возможностью лёгкого добавления новых источников новостей. Для добавления нового сайта:

    - Реализуйте функцию для парсинга новостей с нового сайта.

    - Добавьте новый сайт в список обрабатываемых сайтов

- ***Поддержка высокой нагрузки***
    - Благодаря многопоточности и параллельной обработке запросов API, приложение способно поддерживать высокий уровень запросов и нагрузку на сервер.

    - База данных PostgreSQL используется для хранения и поиска новостей, что обеспечивает надёжность и эффективность при работе с большими объёмами данных.