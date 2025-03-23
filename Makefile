.PHONY: build run validate clean

# ビルド
build:
	go build -o bin/batch cmd/batch/main.go

# 実行
run: build
	./bin/batch

# バリデーション
validate:
	go mod tidy
	go vet ./...
	golangci-lint run

##
# Linux向けクロスコンパイル
##
build-linux:
	@echo "==> Cross compiling for Linux (amd64)"
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 make build

# クリーンアップ
clean:
	rm -rf bin/ 