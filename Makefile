.PHONY: all build test clean validate install-tools

# ビルド後の出力先ディレクトリ
BUILD_DIR     = bin

# allターゲットでは「validate → build → run」を一括実行
all: validate build run

# ビルド
build:
	go build -ldflags "-s -w" -o $(BUILD_DIR)/reservation-batch cmd/batch/reservation/main.go

# クリーンアップ
clean:
	@echo "==> Cleaning build outputs"
	rm -rf $(BUILD_DIR)/

##
# 検証系: fmt, vet, test
##
validate:
	@echo "==> Running go fmt"
	go fmt ./...

	@echo "==> Running go vet"
	go vet ./...

	@echo "==> Running golangci-lint"
	golangci-lint run

#   必要に応じてテストを実行する
# 	@echo "==> Running tests"
# 	go test -v ./...

##
# 実行
##
run:
	@echo "==> Running..."
	@$(BUILD_DIR)/reservation-batch

##
# テスト (validate でも実行しているが、個別でも呼び出せるように)
##
test:
	@echo "==> Running tests"
	go test -v ./...

# 開発ツールのインストール
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

##
# 依存関係の更新: go mod tidy
##
update-deps:
	@echo "==> Updating dependencies"
	go mod tidy

##
# Linux向けクロスコンパイル
##
build-linux:
	@echo "==> Cross compiling for Linux (amd64)"
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 make build 