# go_week1 — wordstat

Небольшой учебный проект: считаем частоты слов в тексте.

В репозитории два бинаря:

- **`wordstat`** — CLI (Command Line Interface) утилита: читает текст из stdin или файлов, печатает `word count`.
- **`wordstatd`** — HTTP server: принимает текст в POST body и отдаёт статистику (text или json), поддерживает health-check, метрики и (опционально) pprof (profiling) на localhost.

---

## Требования

- Go (версия из `go.mod`)
- Windows / Linux / macOS

---

## Структура

- `cmd/wordstat` — CLI entrypoint
- `cmd/wordstatd` — HTTP server entrypoint
- `cmd/internal/wordstat` — основная логика + тесты/бенчмарки

---

## Сборка

```bash
go build ./cmd/wordstat
go build ./cmd/wordstatd
