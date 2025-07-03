# KoeMoji-Go Makefile
# バージョン管理と開発タスクの自動化

# バージョン情報をversion.goから動的に取得
VERSION := $(shell grep -o 'const Version = "[^"]*"' version.go | cut -d'"' -f2)

.PHONY: help version bump-version build build-macos build-windows clean test lint fmt

# デフォルトターゲット
help:
	@echo "KoeMoji-Go 開発用 Makefile"
	@echo ""
	@echo "利用可能なコマンド:"
	@echo "  version          現在のバージョンを表示"
	@echo "  bump-version     バージョンを更新 (使用法: make bump-version NEW_VERSION=1.6.0)"
	@echo "  build            全プラットフォーム向けビルド"
	@echo "  build-macos      macOS向けビルド"
	@echo "  build-windows    Windows向けビルド"
	@echo "  test             テスト実行"
	@echo "  clean            ビルド成果物をクリーンアップ"
	@echo "  fmt              Go コードフォーマット"
	@echo "  lint             Go コードリント"
	@echo ""
	@echo "現在のバージョン: $(VERSION)"

# バージョン情報表示
version:
	@echo "$(VERSION)"

# バージョン更新
bump-version:
ifndef NEW_VERSION
	$(error NEW_VERSION が指定されていません。使用法: make bump-version NEW_VERSION=1.6.0)
endif
	@echo "バージョンを $(VERSION) から $(NEW_VERSION) に更新します..."
	@./scripts/update-version.sh $(NEW_VERSION)
	@echo "✅ バージョン更新完了: $(NEW_VERSION)"

# 全プラットフォーム向けビルド
build: build-macos build-windows

# macOS向けビルド
build-macos:
	@echo "🍎 macOS向けビルド中..."
	@cd build/macos && ./build.sh
	@echo "✅ macOS ビルド完了"

# Windows向けビルド（WSL/Linux環境想定）
build-windows:
	@echo "🪟 Windows向けビルド中..."
	@cd build/windows && ./build.bat
	@echo "✅ Windows ビルド完了"

# 開発用ビルド（現在のプラットフォーム向け）
build-dev:
	@echo "🔧 開発用ビルド中..."
	@go build -o koemoji-go ./cmd/koemoji-go
	@echo "✅ 開発用ビルド完了: koemoji-go"

# テスト実行
test:
	@echo "🧪 テスト実行中..."
	@go test ./...
	@echo "✅ テスト完了"

# ベンチマーク実行
bench:
	@echo "⚡ ベンチマーク実行中..."
	@go test -bench=. ./...

# Go コードフォーマット
fmt:
	@echo "📝 Go コードフォーマット中..."
	@go fmt ./...
	@echo "✅ フォーマット完了"

# Go コードリント（golangci-lintが必要）
lint:
	@echo "🔍 Go コードリント中..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint がインストールされていません"; \
		echo "インストール: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# ビルド成果物のクリーンアップ
clean:
	@echo "🧹 クリーンアップ中..."
	@rm -f koemoji-go
	@rm -rf build/macos/dist
	@rm -rf build/windows/dist
	@rm -rf build/releases
	@echo "✅ クリーンアップ完了"

# Git関連操作
git-status:
	@git status

git-add:
	@git add .

# バージョンタグ作成
create-tag:
	@echo "📋 Git タグ v$(VERSION) を作成中..."
	@git tag v$(VERSION)
	@echo "✅ タグ作成完了: v$(VERSION)"
	@echo "リモートにプッシュ: git push origin v$(VERSION)"

# リリース準備（バージョン更新 + タグ作成）
prepare-release:
ifndef NEW_VERSION
	$(error NEW_VERSION が指定されていません。使用法: make prepare-release NEW_VERSION=1.6.0)
endif
	@make bump-version NEW_VERSION=$(NEW_VERSION)
	@git add .
	@git commit -m "chore: bump version to $(NEW_VERSION)"
	@make create-tag
	@echo ""
	@echo "🎉 リリース準備完了!"
	@echo ""
	@echo "次のステップ:"
	@echo "  1. プッシュ: git push origin feature/version-automation"
	@echo "  2. タグプッシュ: git push origin v$(NEW_VERSION)"
	@echo "  3. プルリクエスト作成"
	@echo ""

# 依存関係更新
deps-update:
	@echo "📦 Go モジュール依存関係更新中..."
	@go mod tidy
	@go mod download
	@echo "✅ 依存関係更新完了"

# 開発環境セットアップ
setup-dev:
	@echo "🛠️  開発環境セットアップ中..."
	@go mod tidy
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "📥 golangci-lint インストール中..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "✅ 開発環境セットアップ完了"