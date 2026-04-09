# mak

`mak` は Makefile のターゲットを fuzzy find で検索し、変数を対話的に入力して
`make` コマンドを実行する TUI ツールです。

## 特徴

- ターゲットの fuzzy 検索（前方一致・連続一致を優先）
- `## ` コメントをターゲットの説明文として表示
- レシピ中の `$(VAR)` / `${VAR}` を自動抽出して対話的に入力
- 実行前にコマンドをプレビュー
- `Ctrl+Y` でコマンドをクリップボードにコピー

## インストール

```sh
go install github.com/doskoiyuta/mak@latest
```

またはリポジトリをクローンしてビルド:

```sh
git clone https://github.com/doskoiyuta/mak.git
cd mak
make build
./mak
```

## 使い方

```
mak [オプション] [Makefileパス]
```

| 引数 | デフォルト | 説明 |
|------|-----------|------|
| `Makefileパス` | `./Makefile` | 対象の Makefile パス |
| `-f <file>` | `./Makefile` | Makefile を明示指定 |
| `-C <dir>` | `.` | 実行ディレクトリを指定 |
| `--dry-run` | `false` | コマンドを表示するだけで実行しない |
| `--version` | — | バージョンを表示 |

## キーバインド

| キー | 動作 |
|------|------|
| 文字入力 | fuzzy 検索 |
| `↑` / `↓` | ターゲット選択 |
| `Tab` | 変数入力欄へフォーカス移動 |
| `Shift+Tab` | 検索欄へ戻る |
| `Enter` | 実行（変数欄では次の変数へ、最後では実行） |
| `Ctrl+Y` | コマンドをクリップボードにコピーして終了 |
| `Ctrl+C` / `Esc` / `q` | 終了 |
| `?` | ヘルプ表示 |

## Makefile の書き方

ターゲット直前の `## ` コメントが説明文として表示されます。

```makefile
APP_NAME = myapp
VERSION ?= 1.0.0

## build: アプリをビルドする
build:
	go build -o $(APP_NAME) -ldflags "-X main.Version=$(VERSION)" ./...
```

レシピ中の `$(APP_NAME)` / `$(VERSION)` は自動的に変数入力欄として表示され、
トップレベルの変数定義（`=`, `?=`, `:=`）はデフォルト値として使用されます。

## 開発

```sh
make all       # fmt, vet, lint, test, build
make test      # go test -race
make lint      # golangci-lint
```

Go のバージョンは `.go-version` に記載されています（goenv 対応）。
