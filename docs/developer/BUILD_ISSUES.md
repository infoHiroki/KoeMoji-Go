# ビルド問題について

## GitHub Actions でのビルド無効化

### 背景
KoeMoji-Go は PortAudio を使用した録音機能を持つため、CGO（C言語バインディング）を必要とします。
これにより、GitHub Actions でのクロスコンパイルが困難になりました。

### 問題点
1. **Windows ビルド**: CGO_ENABLED=1 が必要なため、クロスコンパイルができない
2. **macOS ビルド**: 単一プラットフォームのみのビルドは CI/CD の価値が限定的

### 現在の対応
- GitHub Actions でのビルドワークフローを削除
- 各プラットフォームでのローカルビルドに移行

### ビルド方法

#### macOS
```bash
cd build/macos
./build.sh
```

#### Windows (MSYS2/MinGW64 環境)
```bash
cd build/windows
build.bat
```

### 将来的な改善案
1. CGO 不要な代替ライブラリへの移行
2. プラットフォーム固有のランナーを使用した self-hosted runners
3. 録音機能を別プロセスとして分離