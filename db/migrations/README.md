# Local dev env

### add

```bash
migrate create -ext sql -dir db/migrations -seq ${comments}
```
### migrate

```bash
migrate -path ./db/migrations -database postgres://skyline:skyline@localhost:5432/skyline?sslmode=disable up
```
