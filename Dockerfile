# ビルドステージ
FROM golang:1.21-alpine AS builder

# 必要なパッケージのインストール
RUN apk add --no-cache git

# 作業ディレクトリの設定
WORKDIR /app

# 依存関係のコピーとダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# アプリケーションのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/batch cmd/batch/main.go

# 実行ステージ
FROM alpine:latest

# タイムゾーンの設定
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    echo "Asia/Tokyo" > /etc/timezone && \
    apk del tzdata

# 必要なパッケージのインストール
RUN apk add --no-cache ca-certificates

# 作業ディレクトリの設定
WORKDIR /app

# ビルドしたバイナリのコピー
COPY --from=builder /app/batch .

# 設定ファイルのコピー
COPY config/config.yaml ./config/

# 実行ユーザーの設定
RUN adduser -D -g '' appuser
USER appuser

# アプリケーションの実行
CMD ["./batch"] 