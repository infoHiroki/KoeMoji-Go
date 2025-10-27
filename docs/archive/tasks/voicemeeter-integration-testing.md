# VoiceMeeter統合機能 - Windows環境テストタスク

## 概要
VoiceMeeter統合機能（音声正規化含む）のWindows環境での実動作テストタスク。

**対象ブランチ**: `feature/voicemeeter-integration`

**実装完了日**: 2025年（実装済み、テスト待ち）

## 実装済み内容

### Phase 1: VoiceMeeter検出と音声正規化
- ✅ VoiceMeeter Output/Inputデバイスの自動検出
- ✅ 音声正規化エンジン（閾値5000、目標20000）
- ✅ GUI設定ダイアログへの「VoiceMeeter設定を適用」ボタン追加
- ✅ 音量自動調整チェックボックス追加
- ✅ 包括的なユニットテスト（Mac環境でパス確認済み）
- ✅ GUI動作確認（Mac環境でエラーハンドリング確認済み）

### 技術詳細
- **検出キーワード**: "voicemeeter output", "voicemeeter input", "voicemeeter aux", "cable output"
- **正規化ロジック**: 最大振幅が5000未満の場合のみ20000まで増幅
- **クリッピング保護**: int16範囲（±32767）を超えないよう制限
- **設定フィールド**: `audio_normalization_enabled` (デフォルト: `true`)

## Windows環境テスト項目

### 前提条件
- [ ] Windows 10/11 環境
- [ ] VoiceMeeter（Banana/Potato含む）インストール済み
- [ ] KoeMoji-Go ビルド済み（feature/voicemeeter-integrationブランチ）

### 1. VoiceMeeter検出テスト

#### 1.1 VoiceMeeter起動状態での検出
- [ ] VoiceMeeterを起動
- [ ] KoeMoji-Goを起動
- [ ] 設定ダイアログを開く
- [ ] 「VoiceMeeter設定を適用」ボタンをクリック
- [ ] **期待結果**:
  - ✅ VoiceMeeter Outputデバイスが自動選択される
  - ✅ 音量自動調整がONになる
  - ✅ 成功ダイアログが表示される（デバイス名を含む）

#### 1.2 VoiceMeeter未起動/未インストール時の検出
- [ ] VoiceMeeterを終了（またはアンインストール状態）
- [ ] 「VoiceMeeter設定を適用」ボタンをクリック
- [ ] **期待結果**:
  - ✅ 「VoiceMeeterが見つかりません」ダイアログが表示される
  - ✅ ガイダンスメッセージが表示される

### 2. 音声正規化機能テスト

#### 2.1 低音量録音の正規化
- [ ] VoiceMeeter設定を適用
- [ ] 音量自動調整をON
- [ ] 意図的に低音量で録音（VoiceMeeterのゲインを下げる）
- [ ] 録音停止・保存
- [ ] **期待結果**:
  - ✅ 保存されたWAVファイルの音量が適切に増幅されている
  - ✅ ログに正規化適用メッセージが記録される

#### 2.2 通常音量録音の非変更
- [ ] 通常の音量設定で録音
- [ ] 録音停止・保存
- [ ] **期待結果**:
  - ✅ 音量が変更されない（閾値5000以上のため）
  - ✅ クリッピングが発生しない

#### 2.3 正規化OFF時の動作
- [ ] 音量自動調整をOFF
- [ ] 低音量で録音
- [ ] **期待結果**:
  - ✅ 音量が増幅されない（元の音量のまま保存）

### 3. 統合フローテスト

#### 3.1 初回セットアップフロー
- [ ] 新規インストール状態からスタート
- [ ] VoiceMeeterインストール・設定
- [ ] KoeMoji-Go起動
- [ ] 「VoiceMeeter設定を適用」でデバイス選択
- [ ] テスト録音
- [ ] **期待結果**:
  - ✅ スムーズなセットアップ体験
  - ✅ 録音音量が適切

#### 3.2 デバイス変更後の動作
- [ ] 別の録音デバイスを手動選択
- [ ] 再度「VoiceMeeter設定を適用」
- [ ] **期待結果**:
  - ✅ VoiceMeeterデバイスに上書きされる
  - ✅ 正規化設定が有効になる

### 4. エッジケーステスト

#### 4.1 複数VoiceMeeterデバイス検出
- [ ] VoiceMeeter BananaまたはPotatoを使用（複数の仮想デバイス）
- [ ] 検出動作を確認
- [ ] **期待結果**:
  - ✅ いずれかのVoiceMeeterデバイスが選択される

#### 4.2 長時間録音での正規化
- [ ] 長時間録音（30分以上）を実施
- [ ] 正規化処理時間を計測
- [ ] **期待結果**:
  - ✅ 正規化処理が完了する
  - ✅ メモリエラーが発生しない

#### 4.3 config.json直接編集
- [ ] `audio_normalization_enabled: false` に設定
- [ ] KoeMoji-Go再起動
- [ ] **期待結果**:
  - ✅ チェックボックスがOFFになっている
  - ✅ 正規化が無効化される

## テスト結果記録

### 実施日時
- 未実施

### テスト環境
- OS: Windows 10/11
- VoiceMeeterバージョン:
- KoeMoji-Goバージョン:
- ブランチ: `feature/voicemeeter-integration`

### 結果サマリー
| テスト項目 | 結果 | 備考 |
|-----------|------|------|
| VoiceMeeter検出（起動時） | ⬜ 未実施 | |
| VoiceMeeter検出（未検出時） | ⬜ 未実施 | |
| 低音量録音の正規化 | ⬜ 未実施 | |
| 通常音量の非変更 | ⬜ 未実施 | |
| 正規化OFF時の動作 | ⬜ 未実施 | |
| 初回セットアップフロー | ⬜ 未実施 | |
| デバイス変更後の動作 | ⬜ 未実施 | |
| 複数デバイス検出 | ⬜ 未実施 | |
| 長時間録音 | ⬜ 未実施 | |
| config.json直接編集 | ⬜ 未実施 | |

### 発見された問題
なし（未実施）

### 次のアクション
- Windows環境でのテスト実施
- 問題があれば修正
- テスト完了後、mainブランチへのマージ検討

## 参考ドキュメント
- 設計書: `docs/design/VoiceMeeterIntegration.md`
- ユーザー向けガイド: `docs/user/RECORDING_SETUP.md`
- 実装コミット: `feature/voicemeeter-integration`ブランチ参照
