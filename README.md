# urlshortener

Сервис сокращения URL.

## Сборка сервиса

При сборке можно передать информацию о версии через флаги `-ldflags`:

```bash
go build -ldflags="-X 'github.com/ASTeterin/urlshortener/cmd/shortener.buildVersion=1.0.0' -X 'github.com/ASTeterin/urlshortener/cmd/shortener.buildDate=2023-10-27' -X 'github.com/ASTeterin/urlshortener/cmd/shortener.buildCommit=abc123'" -o ./bin/shortener ./cmd/shortener
```

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**
