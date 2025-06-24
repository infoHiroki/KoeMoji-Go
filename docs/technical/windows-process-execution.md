# Windows実行ファイルの外部プロセス起動に関する技術ノート

## 概要

Windows環境でGUIアプリケーションから外部プロセスを起動する際の、コンソールウィンドウ表示問題とその解決方法について記述します。

## 問題

Goで`-H=windowsgui`フラグを使用してビルドしたGUIアプリケーションでも、`exec.Command`で外部プロセスを起動すると、デフォルトでコンソールウィンドウが表示されます。

## 解決方法

### 基本的なアプローチ

`syscall.SysProcAttr`を使用して、Windowsプロセス作成フラグを設定：

```go
cmd := exec.Command(name, args...)
cmd.SysProcAttr = &syscall.SysProcAttr{
    HideWindow: true,
    CreationFlags: 0x08000000, // CREATE_NO_WINDOW
}
```

### 例外処理が必要なケース

#### explorer.exe

Windowsのファイルエクスプローラー（`explorer.exe`）は、上記のフラグと互換性がありません。`HideWindow`フラグを設定すると、エクスプローラーが起動しない、または正しく動作しません。

**解決策**: `explorer.exe`に対しては通常の`exec.Command`を使用

```go
func createCommand(name string, args ...string) *exec.Cmd {
    cmd := exec.Command(name, args...)
    
    // explorer.exeは例外処理
    if name != "explorer" && name != "explorer.exe" {
        cmd.SysProcAttr = &syscall.SysProcAttr{
            HideWindow: true,
            CreationFlags: 0x08000000,
        }
    }
    
    return cmd
}
```

## ビルド制約を使用した実装

プラットフォーム固有のコードを分離：

**exec_windows.go**:
```go
//go:build windows
// +build windows

package ui

import (
    "os/exec"
    "syscall"
)

func createCommand(name string, args ...string) *exec.Cmd {
    cmd := exec.Command(name, args...)
    if name != "explorer" && name != "explorer.exe" {
        cmd.SysProcAttr = &syscall.SysProcAttr{
            HideWindow: true,
            CreationFlags: 0x08000000,
        }
    }
    return cmd
}
```

**exec_other.go**:
```go
//go:build !windows
// +build !windows

package ui

import "os/exec"

func createCommand(name string, args ...string) *exec.Cmd {
    return exec.Command(name, args...)
}
```

## 適用範囲

この手法は以下のコマンドで有効：
- `pip install`
- `whisper-ctranslate2`
- `notepad.exe`
- その他の一般的なコマンドラインツール

## 注意事項

1. **互換性テスト**: 新しい外部プログラムを追加する際は、フラグとの互換性をテストする必要があります
2. **エラーハンドリング**: プロセス起動に失敗した場合のエラーハンドリングを適切に実装
3. **ログ記録**: デバッグのため、プロセス起動の成功/失敗をログに記録

## 参考資料

- [Windows Process Creation Flags](https://docs.microsoft.com/en-us/windows/win32/procthread/process-creation-flags)
- [Go syscall package](https://pkg.go.dev/syscall)
