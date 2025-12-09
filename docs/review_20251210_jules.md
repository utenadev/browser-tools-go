# Code Review for browser-tools-go (2025-12-10)

## Summary

This review covers the initial state of the `browser-tools-go` codebase, primarily focusing on the `main.go` file. The tool is a powerful command-line utility for browser automation, but it has several areas that could be improved in terms of structure, reliability, and maintainability.

## Key Findings

### 1. Monolithic Structure
- **Issue:** All the application logic, command definitions, and data structures are located in a single `main.go` file.
- **Impact:** This makes the code difficult to navigate, maintain, and test. It also hinders code reuse.
- **Recommendation:** Break the code into multiple files and packages. For example:
    - `cmd/`: Cobra command definitions.
    - `internal/browser/`: Browser context management (`newPersistentContext`, `newTemporaryContext`).
    - `internal/logic/`: Core application logic for each command (`logicNavigate`, `logicScreenshot`, etc.).
    - `internal/models/`: Struct definitions (`SearchResult`, `HnSubmission`, etc.).

### 2. Lack of Unit Tests
- **Issue:** The project has no automated tests.
- **Impact:** Refactoring the code or adding new features is risky, as there's no way to verify that existing functionality hasn't broken.
- **Recommendation:** Introduce unit tests for the business logic. The `chromedp` calls can be mocked or interfaced to allow for testing without running a real browser instance.

### 3. Use of `time.Sleep` for Waiting
- **Issue:** The code frequently uses `time.Sleep(2 * time.Second)` to wait for web pages to load or for asynchronous operations to complete.
- **Impact:** This is unreliable. The sleep duration might be too short for slow networks or too long for fast ones, leading to either flaky tests or inefficient execution.
- **Recommendation:** Replace `time.Sleep` with `chromedp`'s built-in waiting functions, such as `chromedp.WaitVisible`, `chromedp.WaitReady`, or `chromedp.Poll`.

### 4. Inconsistent Error Handling
- **Issue:** Error handling is inconsistent. Many functions use `log.Fatalf`, which abruptly terminates the entire application.
- **Impact:** This makes the tool inflexible. A failure in one command (e.g., a timeout) shouldn't necessarily crash the whole program. It also makes the logic functions impossible to test without the test process exiting.
- **Recommendation:** Logic functions should return errors to their callers. The command execution layer (in `Run` functions) should be responsible for deciding how to present the error to the user (e.g., printing to `stderr` and exiting with a non-zero status code).

### 5. Redundant `runCmd` Implementation
- **Issue:** The `run` command manually re-parses arguments and flags to execute a subcommand in a temporary browser.
- **Impact:** This is complex and brittle. Adding a new flag to a subcommand requires updating the `run` command's flag set as well, which is easy to forget.
- **Recommendation:** Refactor the command structure. A better approach would be to use a persistent pre-run function that manages the browser context (either persistent or temporary) based on a global flag (e.g., `--run-in-temp-browser`).

### 6. Hardcoded CSS Selectors
- **Issue:** CSS selectors for scraping Google and Hacker News are hardcoded within the logic functions.
- **Impact:** If the layout of these websites changes, the tool will break and require a code update.
- **Recommendation:** While moving them to a config file is an option, a more robust solution involves building more resilient selectors or having fallback selectors, as is partially implemented in `logicSearch`.

### 7. Global `rootCmd` Variable
- **Issue:** The `rootCmd` is a global variable.
- **Impact:** This can make testing difficult, as tests might inadvertently affect each other by modifying the same global state.
- **Recommendation:** Encapsulate the command setup in a `newRootCmd()` function that returns a `*cobra.Command`. This makes it possible to create fresh command instances for tests.

### 8. Mixing Logging and Standard Output
- **Issue:** The tool mixes informational logging (`log.Printf`) with result output (`fmt.Println`).
- **Impact:** This makes it difficult for users to pipe the JSON output to other tools (like `jq`) because the log messages will also be captured in `stdout`.
- **Recommendation:** Use the `log` package to write all informational/debug messages to `stderr`. Write the final result (JSON or text) to `stdout`.

## Conclusion

The tool is functional and provides useful features. The recommendations above are intended to improve the codebase's quality, making it easier to maintain, test, and extend in the future. The highest priority should be on splitting the monolithic file and introducing unit tests.

## 更新: リファクタリングによる改善点

このレビューに基づき、以下のリファクタリングを実施しました。

- **1. モノリシックな構造:**
  - **解決済み:** コードを`internal/cmd`, `internal/logic`, `internal/models`, `internal/browser`の各パッケージに分割しました。これにより、コードの見通しとメンテナンス性が向上しました。

- **2. 単体テストの欠如:**
  - **部分的に解決:** `internal/logic`パッケージに単体テストを追加し、コンテンツのフォーマット変換ロジックをテストしました。今後、他のロジックについてもテストを追加することが推奨されます。

- **3. `time.Sleep`の使用:**
  - **解決済み:** 信頼性の低い`time.Sleep`を`chromedp.WaitVisible`に置き換え、より堅牢な待機処理を実装しました。

- **4. エラーハンドリングの不整合:**
  - **解決済み:** `logic`層の関数が`log.Fatalf`を呼び出す代わりにエラーを返すように修正しました。これにより、テスト可能性が向上しました。

- **8. ロギングと出力の混在:**
  - **解決済み:** `log`パッケージからの出力を`stderr`に、コマンドの結果（JSONなど）を`stdout`に出力するように分離しました。これにより、`jq`などのツールとの連携が容易になりました。

### 未解決の項目

- **5. `runCmd`の冗長な実装:**
  - 大規模な変更が必要なため、今回のリファクタリングでは見送りました。
- **7. グローバルな`rootCmd`変数:**
  - `internal/cmd`パッケージ内にカプセル化されましたが、よりテストしやすい構造への改善は今後の課題です。
