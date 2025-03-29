# Echo Playground Batch Task

Echo Playgroundのバッチ処理マイクロサービスです。

## 機能

- 予約バッチ処理
  - 保留中の予約を処理
  - 重複予約のチェック
  - 予約ステータスの更新

## 必要条件

- Go 1.21以上
- PostgreSQL 13以上
- Docker (オプション)

## セットアップ

1. リポジトリのクローン

```bash
git clone https://github.com/horsewin/echo-playground-batch-task.git
cd echo-playground-batch-task
```

2. 依存関係のインストール

```bash
go mod download
```

3. 開発ツールのインストール

```bash
make install-tools
```

## ビルドと実行

### ローカル環境

1. ビルド

```bash
make build
```

2. 実行

```bash
./bin/reservation-batch
```

### Docker環境

1. イメージのビルド

```bash
docker build -t echo-playground-batch-task .
```

2. コンテナの実行

```bash
docker run -e DB_HOST=host.docker.internal \
           -e DB_PORT=5432 \
           -e DB_USER=postgres \
           -e DB_PASSWORD=postgres \
           -e DB_NAME=echo_playground \
           echo-playground-batch-task
```

## 環境変数

| 変数名      | 説明                   | デフォルト値 |
| ----------- | ---------------------- | ------------ |
| DB_HOST     | データベースホスト     | localhost    |
| DB_PORT     | データベースポート     | 5432         |
| DB_USERNAME | データベースユーザー   | sbcntrapp    |
| DB_PASSWORD | データベースパスワード | password     |
| DB_NAME     | データベース名         | sbcntrapp    |

## 開発コマンド

- `make build`: アプリケーションのビルド
- `make test`: テストの実行
- `make validate`: コードの検証
- `make clean`: ビルド成果物の削除
- `make install-tools`: 開発ツールのインストール

## ライセンス

MIT
