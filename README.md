# go_week1 — wordstat

CLI tool that counts word frequencies from stdin.

## Build
```bash
go build ./cmd/wordstat
```

## Run
```bash
echo "aa bb aa" | go run ./cmd/wordstat -sort=count
```

## Flags
```-k``` how many entries to print(0 = all)

```-min``` minimum count to include

```-sort``` sort by: word|count
