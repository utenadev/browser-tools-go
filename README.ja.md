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
