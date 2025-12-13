# receipt-scan-backend

## PostgreSQL を Docker で起動する
```
docker compose up -d db
```

接続先は `DATABASE_URL` で指定します。未設定の場合は以下のデフォルトで接続します。

```
postgres://receipt:receipt@localhost:5432/receipt?sslmode=disable
```

> `docker-compose.yml` は `postgres:16-alpine` ベースで、ユーザー/パスワードともに `receipt`、データベース名 `receipt` で起動します。

## アプリの起動
環境変数を `.env` などで設定した上で実行します。

```
go run ./cmd/api
```
