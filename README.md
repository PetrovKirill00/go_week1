# go_week1 — wordstat

Учебный проект на Go: считаем частоты слов в тексте.

В репозитории два бинаря:

- **`wordstat`** — CLI (Command Line Interface) утилита: читает текст из stdin или файлов, печатает статистику.
- **`wordstatd`** — HTTP server: принимает текст в POST body и отдаёт статистику (text/json), поддерживает health-check, метрики, request id, recovery, graceful shutdown и (опционально) pprof (profiling) на localhost.

---

## Quickstart

### 1) Тесты
```bash
go test ./...
```

### 2) CLI
```bash
echo "b a a b c" | go run ./cmd/wordstat -sort=count
```

### 3) HTTP server (порт 8080)
```bash
go run ./cmd/wordstatd -addr :8080
```

Проверка:
```powershell
curl.exe "http://localhost:8080/healthz"
curl.exe -X POST "http://localhost:8080/wordstat?sort=count&format=json" -d "b a a b c"
```

> На Windows PowerShell `curl` часто является алиасом для `Invoke-WebRequest`, поэтому используй именно `curl.exe`.

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
```

Опционально собрать в `bin/`:

```bash
go build -o bin/wordstat  ./cmd/wordstat
go build -o bin/wordstatd ./cmd/wordstatd
```

---

## CLI: `wordstat`

### Использование
```bash
wordstat [flags] [files...]
```
Если `files` не указаны — читает из stdin.

### Флаги (основные)
- `-sort` — `word|count`
- `-k` — сколько строк вывести (`0` = все)
- `-min` — минимальный count, чтобы слово попало в вывод
- (если есть) `-format` — `text|json`

Актуальный список:
```bash
go run ./cmd/wordstat -h
```

### Примеры

stdin:
```bash
echo "aa bb aa" | go run ./cmd/wordstat -sort=count
```

файлы:
```bash
go run ./cmd/wordstat -sort=count f1.txt f2.txt
```

---

## HTTP server: `wordstatd`

### Запуск
```bash
go run ./cmd/wordstatd -addr :8080
```

### Health-check
`GET /healthz` → `200 OK`
```powershell
curl.exe "http://localhost:8080/healthz"
```

---

## API

### `POST /wordstat`

Тело запроса (body): произвольный текст. Слова разделяются пробелами/переносами строк.

Query параметры:

| param   | type  | default | allowed        | meaning |
|--------|-------|---------|----------------|---------|
| `sort` | string| `word`  | `word`,`count` | сортировка |
| `format` | string | `text` | `text`,`json` | формат успешного ответа |
| `k`    | int   | `0`     | `>=0`          | top-k (`0` = все) |
| `min`  | int   | `1`     | `>0`           | минимальный count |

Пример (json):
```powershell
curl.exe -X POST "http://localhost:8080/wordstat?sort=count&format=json" -d "b a a b c"
```

Пример (text):
```powershell
curl.exe -X POST "http://localhost:8080/wordstat?sort=count&format=text" -d "b a a b c"
```

### Формат ответов

#### Успех
- `format=text`: строки вида
  ```
  a 2
  b 2
  c 1
  ```
- `format=json`: формат зависит от реализации (map или array), главное — пары word/count.

#### Ошибки (всегда JSON)
Ошибки всегда возвращаются в JSON:
```json
{"error":"...","request_id":"..."}
```

В каждом ответе сервер ставит заголовок:
- `X-Request-Id: <id>`

Если клиент прислал `X-Request-Id`, сервер использует его. Иначе — генерирует.

---

## Флаги сервера

Список:
```bash
go run ./cmd/wordstatd -h
```

Обычно есть:

- `-addr` — адрес основного сервера (например `:8080`)
- `-max-body` — лимит POST body в bytes (байтах)
- `-read-timeout`, `-write-timeout` — таймауты чтения/записи
- `-shutdown-timeout` — время на graceful shutdown
- `-dev` — dev-mode: отключает read/write timeouts (удобно для slow-client тестов)
- `-pprof` — включить pprof server (только localhost)
- `-pprof-addr` — адрес pprof сервера (например `127.0.0.1:6060`)

Пример запуска без таймаутов для проверки медленного клиента:
```powershell
go run ./cmd/wordstatd -addr :8080 -dev
```

Slow client тест (искусственно медленная отправка body):
```powershell
python -c "print('a ' * 40000, end='')" > big.txt
curl.exe --limit-rate 10k --data-binary "@big.txt" "http://localhost:8080/wordstat?sort=count&format=json"
```

---

## Debug endpoints

### expvar (метрики)
`GET /debug/vars` на основном сервере:
```powershell
curl.exe "http://localhost:8080/debug/vars"
```

### pprof (profiling)
Если сервер запущен с `-pprof`, pprof доступен на localhost:
```powershell
curl.exe "http://127.0.0.1:6060/debug/pprof/"
```

Снять CPU profile:
```bash
go tool pprof "http://127.0.0.1:6060/debug/pprof/profile?seconds=10"
```

---

## Middleware (что включено)

- `RequestID` — выставляет/пробрасывает `X-Request-Id`
- `Logging` — логирует method/path/status/bytes/duration + req_id
- `Recovery` — ловит panic, возвращает 500 и логирует stack trace

Порядок важен:
- `RequestID(Logging(Recovery(mux)))`

---

## Graceful shutdown

Сервер корректно завершается по Ctrl+C (SIGINT (signal interrupt)):

- перестаёт принимать новые запросы
- даёт текущим запросам ограниченное время завершиться (`-shutdown-timeout`)
- затем закрывается

---

## Тесты

```bash
go test ./...
```

---

## Benchmarks

```bash
go test -run ^$ -bench . -benchmem ./cmd/internal/wordstat
```

С фиксированным временем и повторениями:
```bash
go test -run ^$ -bench BenchmarkCount -benchmem -benchtime=2s -count 5 ./cmd/internal/wordstat
```

CPU / memory profiles:
```bash
go test -run ^$ -bench BenchmarkCountBuffered -benchtime=2s -count 3 -cpuprofile cpu.pprof ./cmd/internal/wordstat
go tool pprof -top cpu.pprof

go test -run ^$ -bench BenchmarkCountBuffered -benchtime=2s -count 3 -memprofile mem.pprof -benchmem ./cmd/internal/wordstat
go tool pprof -top mem.pprof
```

---

## Troubleshooting (Windows)

### “Only one usage of each socket address …”
Порт занят. Найти кто слушает 8080:
```powershell
netstat -ano | findstr :8080
```
Убить процесс:
```powershell
taskkill /PID <PID> /F
```

---
