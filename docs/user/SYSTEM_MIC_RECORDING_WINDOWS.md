# Windows環境でシステム音声とマイク音声を同時録音する設定

このドキュメントでは、KoeMoji-GoでPCのシステム音声（YouTube、Zoom等）とマイク音声を同時に録音する方法を説明します。

## 概要

WindowsではmacOSのBlackHoleに相当する機能として、VB-Audio Virtual Cableを使用します。これにより、コードの変更なしでシステム音声とマイク音声の同時録音が可能になります。

## 必要なソフトウェア

- **VB-Audio Virtual Cable**（無料）
  - 仮想オーディオケーブルソフトウェア
  - ダウンロード: https://vb-audio.com/Cable/

## セットアップ手順

### 1. VB-Audio Virtual Cableのインストール

1. [VB-Audio公式サイト](https://vb-audio.com/Cable/)から`VBCABLE_Driver_Pack43.zip`をダウンロード
2. ZIPファイルを解凍
3. **重要**: `VBCABLE_Setup_x64.exe`を**右クリック**→**「管理者として実行」**
4. 「Install Driver」ボタンをクリック
5. Windowsのデバイスソフトウェアインストール確認で「インストール」を選択
6. インストール完了後、**PCを再起動**

### 2. Windowsサウンド設定

#### マイク音声を含める設定
1. サウンドコントロールパネルを開く（`Win+R`→`mmsys.cpl`）
2. 「録音」タブを選択
3. 使用するマイクデバイス（例：Realtek High Definition Audio、USB マイク等）をダブルクリック
4. 「聴く」タブを選択
5. 「このデバイスを聴く」にチェック
6. 「このデバイスを使用して再生する」で**「CABLE Input (VB-Audio Virtual Cable)」**を選択
7. 「OK」をクリック

### 3. KoeMoji-Goの設定

1. KoeMoji-Goを起動
2. GUIモードの場合：設定画面を開く
3. TUIモードの場合：`c`キーを押して設定メニューへ
4. 「録音デバイス」の選択で**「CABLE Output (VB-Audio Virtual Cable)」**を選択
5. 設定を保存

## 動作の仕組み

```
システム音声 → CABLE Input ┐
                          ├→ CABLE Output → KoeMoji-Go（録音）
マイク音声 ───────────────┘
```

1. すべてのシステム音声がCABLE Inputに送られる
2. マイクの「聴く」機能により、マイク音声もCABLE Inputに送られる
3. CABLE InputとCABLE Outputは内部で接続されている
4. KoeMoji-GoはCABLE Outputから両方の音声を録音

## トラブルシューティング

### 音が聞こえない場合
- ヘッドホンやスピーカーを使用したい場合は、Windowsの音量ミキサーで個別に設定
- または、VoiceMeeter Bananaなどのより高機能なソフトを使用

### 録音されない場合
1. VB-Audio Virtual Cableが正しくインストールされているか確認
   - デバイスマネージャーで「サウンド、ビデオ、およびゲームコントローラー」を確認
2. KoeMoji-Goで正しいデバイスが選択されているか確認

### エコーやハウリングが発生する場合
- スピーカーではなくヘッドホンを使用
- マイクの音量を調整

## 他の選択肢

### ステレオミキサー（一部のPCのみ）
一部のWindows PCには「ステレオミキサー」が搭載されています：
1. 録音デバイスで右クリック→「無効なデバイスの表示」
2. ステレオミキサーが表示されたら有効化

### VoiceMeeter Banana（高機能）
より細かい制御が必要な場合：
- 複数の入出力を個別に制御可能
- エフェクトやEQ機能あり
- 設定がより複雑

## 参考情報

- VB-Audio公式サイト: https://vb-audio.com/
- KoeMoji-Go公式リポジトリ: https://github.com/infoHiroki/KoeMoji-Go

---
最終更新: 2024年12月
