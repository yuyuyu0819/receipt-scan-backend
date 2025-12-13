# receipt-scan-backend

## PostgreSQL を Docker で起動する
```
docker compose up -d postgres
```

接続先は `DATABASE_URL` で指定します。未設定の場合は以下のデフォルトで接続します。

```
postgres://receipt:receipt@localhost:5432/receipt?sslmode=disable
```

> `docker-compose.yml` は `postgres:16-alpine` ベースで、ユーザー/パスワードともに `receipt`、データベース名 `receipt` で起動します。

### スキーマの適用
`docker compose` を使っている場合、次のコマンドで `db/schema.sql` を適用できます（サービス名は `postgres`）。

```
cat db/schema.sql | docker compose exec -T postgres psql -U receipt -d receipt
```

ローカルの `psql` を使う場合は次のようにも実行できます。

```
psql "$DATABASE_URL" -f db/schema.sql
```

> 以前のスキーマで `receipts` テーブルに `items` カラム（JSON 形式）が存在する場合、
> `db/schema.sql` にはそのカラムを削除する `ALTER TABLE` が含まれています。
> 上記コマンドを実行することで NOT NULL 制約に起因する「null value in column "items"」エラーを解消できます。

## アプリの起動
環境変数を `.env` などで設定した上で実行します。

```
go run ./cmd/api
```
