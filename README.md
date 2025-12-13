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

### OCR で PDF を送る場合
- リクエストの `imageBase64` に PDF を含めると、サーバー側で自動的に **先頭ページを JPEG に変換** してから Vision API に送ります（Vision のエンドポイントは JPEG/PNG などの画像形式に対応するため）。
- 複数ページ PDF を渡した場合は 1 ページ目のみが OCR 対象になります。

### レシートに紐づく items を取得する API

`/api/receipts/items` に対して、レシート ID を JSON で POST すると紐づく items の配列が返ります。

#### リクエストに必要な情報
- HTTP メソッド: `POST`
- ヘッダー: `Content-Type: application/json`
- ボディ: 次の JSON を送ります。
  - `receiptId` (number, 必須): 取得したいレシートの ID。0 以下はエラーになります。

リクエスト例:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"receiptId":123}' \
  http://localhost:8080/api/receipts/items
```

レスポンス例:

```json
{
  "items": [
    {"id": 1, "receiptId": 123, "name": "りんご", "price": 120},
    {"id": 2, "receiptId": 123, "name": "バナナ", "price": 180}
  ]
}
```
