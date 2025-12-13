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

### 接続情報を変えたい場合
`DATABASE_URL` が空のときは次の環境変数から DSN を組み立てます（未指定はカッコ内の値）。

* `PGHOST` (`localhost`)
* `PGPORT` (`5432`)
* `PGUSER` (`receipt`)
* `PGPASSWORD` (`receipt`)
* `PGDATABASE` (`receipt`)
* `PGSSLMODE` (`disable`)

パスワード認証エラーが出る場合は、実際の DB に設定したパスワードと `PGPASSWORD`/`DATABASE_URL` の値が一致しているかを確認してください。Docker ボリュームに古い資格情報が残っている場合は、`docker compose down -v` でボリュームを削除してから再起動すると初期パスワード（`receipt`）で接続できるようになります。

## アプリの起動
環境変数を `.env` などで設定した上で実行します。

```
go run ./cmd/api
```
