# Browser Tools Go

これは [utena/browser-tools](https://github.com/utenadev/browser-tools) のGoによる再実装です。元のプロジェクトは [badlogic/browser-tools](https://github.com/badlogic/browser-tools) からフォークされました。

エージェント支援型Web自動化のためのGoネイティブなChrome DevTools Protocolツールです。これらのツールは、リモートデバッグが有効な状態で実行されているChromeインスタンスに接続します。

## インストール

```bash
go install github.com/user/browser-tools-go@latest
```
*(注意: `github.com/user/browser-tools-go` は実際のレポジトリパスに置き換えてください)*

これにより、`browser-tools-go` コマンドが `$GOPATH/bin` ディレクトリにインストールされます。

## ヘルプ

```bash
browser-tools-go --help          # 全てのコマンドを表示
browser-tools-go start --help    # start コマンドのオプションを表示
browser-tools-go <command> --help # 特定のコマンドのヘルプを表示
```

## ツールの呼び出し方

**エージェント向け重要事項**: サブコマンドと共に `browser-tools-go` コマンドを使用してください。

✓ 正しい例:
```bash
browser-tools-go start
browser-tools-go navigate https://example.com
browser-tools-go pick "#submit-button"
```

✗ 間違った例:
```bash
./browser-tools-go start  # PATHが通っていれば不要
```

## セッション管理

実行中のブラウザへの接続は自動的に管理されます。

- `start` はChromeを起動し、接続情報を保存します。
- `close` はChromeインスタンスを終了し、接続情報をクリーンアップします。
- その他の全てのコマンドは、保存された接続情報を自動的に使用します。

### Chromeの起動

```bash
browser-tools-go start              # 新しいプロファイルで起動
browser-tools-go start --headless   # ヘッドレスモードで実行
```

リモートデバッグが有効な状態でChromeを起動します。

### Chromeの終了

```bash
browser-tools-go close
```
`start`によって起動されたChromeインスタンスを終了します。

## コマンド

### Navigate (ナビゲート)

```bash
browser-tools-go navigate https://google.com
```

現在のタブを新しいURLに移動します。

### Screenshot (スクリーンショット)

```bash
browser-tools-go screenshot my-shot.png
browser-tools-go screenshot my-shot.png --url https://example.com
browser-tools-go screenshot --url https://example.com --full-page
```

スクリーンショットをキャプチャします。パスが省略された場合、一時ファイルに保存されます。
- `--url <url>`: スクリーンショットを撮る前に指定のURLに移動します。
- `--full-page`: ページ全体をキャプチャします。

### Pick Elements (要素の選択)

```bash
browser-tools-go pick "#submit-button"
browser-tools-go pick ".item-class" --all
```

CSSセレクタに一致する要素の情報を選択・抽出します。
- **`<selector>`**: マッチさせるCSSセレクタ。
- **`--all`**: 最初に見つかった要素だけでなく、一致する全ての要素から情報を抽出します。

### Evaluate JavaScript (JavaScriptの評価)

```bash
browser-tools-go eval 'document.title'
browser-tools-go eval 'document.querySelectorAll("a").length'
```

アクティブなタブでJavaScript式を実行します。コードは非同期コンテキストで実行されます。結果はJSON形式で返されます。

### Cookies (クッキー)

```bash
browser-tools-go cookies
```

現在のブラウザコンテキストの全てのクッキーを表示します。ドメイン、パス、httpOnly、secureフラグも含まれます。

### Search Google (Google検索)

```bash
browser-tools-go search "rust programming"
browser-tools-go search "climate change" --n 10
browser-tools-go search "machine learning" --n 3 --content
```

Googleを検索し、結果を返します。
- `--n <num>`: 返す結果の数（デフォルト: 5）。
- `--content`: 各検索結果のリンクから、読み取り可能なコンテンツ（プレーンテキスト形式）を取得・抽出します。

### Hacker News Scraper (Hacker News スクレイパー)

```bash
browser-tools-go hn-scraper
browser-tools-go hn-scraper 10
```

Hacker Newsのトップ記事をスクレイピングします。取得する記事の数にオプションで制限を設けることができます（デフォルト: 30）。

## スクレイピング信頼性向上

### セレクタフォールバック機構

2025年12月現在、以下の機能が追加され、Webサイトのレイアウト変更に強くなりました：

#### 自動セレクタフォールバック

各スクレイピングコマンドは、以下のように**複数のセレクタ候補**を順番に試します：

**Google検索:**
- コンテナ: `div#search` → `div#rso` → `div.g`
- タイトル: `h3` → `h3.LC20lb` → `div.v9i61e`
- スニペット: `div.VwiC3b` → `div.s` → `div.BNeawe`

**Hacker News:**
- メインテーブル: `table.itemlist` → `table#hnmain` → `table`
- タイトルリンク: `span.titleline > a` → `a.storylink` → `td.title > a`
- スコア: `.score` → `.subtext .score`

最初のセレクタが失敗した場合、自動的に代替セレクタを試します。これにより、Webサイトのレイアウト変更に自動的に対応できます。

### リトライ機構

ネットワークエラーや一時的なエラーに対して、**自動リトライ**が機能します：

```bash
# 以下のようなエラーで一時的に失敗した場合、自動的に最大3回まで再試行
# - ネットワークタイムアウト
# - 接続エラー
# - 一時的なサーバーエラー
```

**リトライ設定:**
- 最大3回のリトライ（設定可能）
- 指数バックオフ（100ms → 200ms → 400ms）
- 最大バックオフ: 2秒

### カスタムセレクタ設定

環境ごとにセレクタをカスタマイズできます：

1. **カスタムセレクタファイルを作成:**

```bash
# 1. 設定ディレクトリを作成
mkdir -p ~/.browser-tools-go

# 2. 設定ファイルを作成
cat > ~/.browser-tools-go/selectors.json << 'EOF'
{
  "google_search": {
    "result_item": ["div.rc", "div.g"],
    "title": ["h3.LC20lb", "h3"],
    "snippet": ["div.VwiC3b", "div.s"]
  },
  "hacker_news": {
    "title_link": ["span.titleline > a", "a.storylink"]
  }
}
EOF
```

2. **効果:**
   - ツールは`selectors.json`に指定されたセレクタを優先して使用
   - フォールバックセレクタが自動的にフォロー
   - Webサイトのレイアウト変更を素早く対応可能

3. **デフォルト設定の確認:**
デフォルト設定を確認するには、以下のコマンドで現在の設定をエクスポートできます：

```bash
# (後で実装可能)
```

### セレクタ推奨リスト

**よく変化するコンポーネントの例:**

| サイト | 脆弱なセレクタ | 代わりのセレクタ |
|--------|--------------|----------------|
| Google | `div.VwiC3b` | `div.s`, `div.BNeawe` |
| HN | `span.titleline` | `a.storylink`, `span.titleline > a` |

### 設定オプション

**selectors.jsonフォーマット:**

```jsonc
{
  "google_search": {
    "search_container": ["div#search", "div#rso"],
    "result_item": ["div.g", "div.rc"],
    "title": ["h3", "h3.LC20lb"],
    "url": ["a[href]"],
    "snippet": ["div.VwiC3b", "div.s"],
    "fallback_wait": ["div#search", "body"]
  },
  "hacker_news": {
    "main_table": ["table.itemlist"],
    "title_link": ["span.titleline > a"],
    "score": [".score"],
    "author": [".hnuser"],
    "time": ["span.age a"],
    "comments": ["td.subtext > a:last-child"],
    "fallback_wait": ["table.itemlist", "body"]
  }
}
```

### 新機能のテスト

新しいスクレイピング機能は以下のテストで検証されています：

```bash
# 新しいユーティリティテスト
go test -v ./internal/utils/selectors_test.go
go test -v ./internal/utils/retry_test.go

# 強化版スクレイパーテスト
go test -v ./internal/logic/scraping_enhanced_test.go
```

### トラブルシューティング

**スクレイピングが頻繁に失敗する場合:**

1. **セレクタ設定の確認:**
```bash
cat ~/.browser-tools-go/selectors.json
```

2. **Chrome DevToolsでの確認:**
```bash
# Chrome DevToolsで現在のDOM構造を確認
# 適切なセレクタを選択し、selectors.jsonに追加
```

3. **リトライ設定のカスタマイズ:**
```json
{
  "retry": {
    "max_attempts": 5,
    "initial_backoff_ms": 1000
  }
}
```

### 今後の予定

- Google検索セレクタ自動更新機能
- セレクタ有効性の自動検出
- Webサイト変更の自動検知と通知

## セキュリティ

### ファイルパス保護

本ツールは、以下のセキュリティ機能を実装して、ファイルパストラバーサル攻撃を防ぎます：

- **SafeWriteFile**: ファイル操作前にパスの安全性を検証します
- **ValidateFilePath**: 危険なファイルパス（`../`、絶対パス、NULLバイトなど）を拒否します
- **スクリーンショット保存**: スクリーンショット出力パスを常に検証し、PNG形式に制限します

#### セキュリティ機能の挙動

以下のパターンはセキュリティにより拒否されます：

```bash
# 親ディレクトリ参照（パストラバーサル）
browser-tools-go screenshot ../secrets.txt          # 拒否: 親ディレクトリ参照
browser-tools-go screenshot ../../../etc/passwd     # 拒否: システムファイルアクセス

# NULLバイト注入
browser-tools-go screenshot "file\x00name.png"       # 拒否: NULLバイト

# 絶対パス（未許可）
browser-tools-go screenshot /tmp/output.png          # 拒否: 絶対パス指定
```

**重要**: クロスサイトスクリプティング（XSS）やコマンドインジェクション攻撃を防ぐため、URLやJSインプットは適切にエスケープされます。HTTPリクエストスプーフィング攻撃を回避するため、URL入力は常にチェックされます。

詳細は[securityYAMLレポート](my/report_20251220_Claude+deepseek.yaml)を確認ください。

### データとプライバシー

- **Cookie情報**: コマンド出力に表示されますが、ローカルAPIの使用に限定されています
- **セッション情報**: `~/.browser-tools-go/ws.json`に保存されます（パーミッション: `0600`）
- **リモート接続**: デフォルトでローカルホスト(`127.0.0.1`)のみを許可しています

## テスト

### 現在のテストカバレッジ

本ツールは、以下のパッケージ単位でテストを実行しています：

| パッケージ | テストカバレッジ | テスト内容 |
|-----------|----------------|----------|
| `internal/utils` | 高 | `ValidateFilePath`, `ValidateScreenshotPath`, `SecureWriteFile`など |
| `internal/config` | 高 | `SaveWsInfo`, `LoadWsInfo`, `RemoveWsInfo`, パーミッション設定 |
| `internal/browser` | 中 | `Start`, `Close`, `WaitForWS`, `NewPersistentContext` |
| `internal/cmd` | 中 | コマンド定義、引数検証、フラグ設定 |
| `internal/logic` | 部分 | `PickElements`, `GetContentFormatting` |

### テストの実行

```bash
# 全パッケージのテストを実行
go test ./...

# 詳細出力でのテスト実行
go test -v ./...

# 特定パッケージのテスト実行
go test -v ./internal/utils
go test -v ./internal/config
go test -v ./internal/browser
go test -v ./internal/cmd
go test -v ./internal/logic

# HTMLカバレッジレポート生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**注意**: Chrome DevToolsとの接続を必要とする完全な統合テストは、Chromeが実行環境にインストールされていないとスキップされる場合があります。

### 最近の改良

2025年12月現在、以下のセキュリティ及びテスト機能強化が実装されました：

1. ✅ **セキュリティ機能**: ファイルパストラバーサル防止（`internal/utils/path.go`）
2. ✅ **セキュリティ機能**: スクリーンショットパス検証強化
3. ✅ **テスト追加**: `config` パッケージ（9種類）
4. ✅ **テスト追加**: `browser` パッケージ（5種類）
5. ✅ **テスト追加**: `cmd` パッケージ（12種類）
6. ✅ **テスト追加**: `utils` パッケージ（17種類）

## ライセンス

[MIT License](LICENSE)
