# ビルドステージ
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 依存関係のコピーとダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# アプリケーションのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/reservation-batch cmd/batch/reservation/main.go

# 実行ステージ
FROM alpine:latest

WORKDIR /app

# ビルドしたバイナリをコピー
COPY --from=builder /app/bin/reservation-batch .

# 実行ユーザーの設定
RUN adduser -D -g '' appuser
USER appuser

# 環境変数の設定
ENV DB_HOST=localhost \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=echo_playground

# エントリーポイントの設定
ENTRYPOINT ["./reservation-batch"] 