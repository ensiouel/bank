## A service for working with the user's balance. Getting the balance, crediting, debiting, transferring.

## Deployment

### docker compose

---

**Build** application

```shell
docker compose build
```

---

**Run** application

```shell
docker compose up -d
```

## Endpoints

### Default port: `8082`

- #### `/api/v1` - REST API
- #### `/swagger` - Documentation

## Config

### All options are loaded from **[.env](.env)**

```dotenv
SERVER_ADDR=:8080

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=bank

APILAYER_APIKEY=apikey
```

## Tests

```
go test ./...
```
