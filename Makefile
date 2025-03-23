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

# クリーンアップ
clean:
	rm -rf bin/ 