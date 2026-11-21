# home-portal

家庭内アプリへの入口と、SwitchBot 機器の最新の観測値を表示するポータルです。
勤怠管理アプリと同じ Go + Vue 3 + TypeScript + Vite + Element Plus の構成とデザイントークンを使います。
Vue のビルド結果を Go に埋め込み、CGO 不要の単一バイナリとして配布します。

## 表示内容

- 勤怠管理 `/kintai/`、料理アプリ `/cooking/` へのリンク
- Meter Plus の温度・湿度・電池残量
- Meter Pro CO₂ の温度・湿度・CO₂濃度・電池残量
- 観測日時、未取得、5分以上更新なし、未来の観測時刻の表示
- Samba 共有先のファイルシステムの空き容量・全体容量・使用済み容量と収集日時

機器名・MAC アドレスとアプリのリンクは [catalog.ts](apps/web/src/catalog.ts) に固定しています。
機器は `home-server/inventory.yml` の2台に合わせています。場所の名前は未設定です。
観測値は `homehub-api` の `/v1/devices` から毎分取得し、画面から再読み込みもできます。
未取得の値をゼロで補わず、取得失敗時は前回の値とエラーを表示します。
一覧編集・認証・永続化用 DB は設けていません。

容量は `home-server` の `disk.yml` が5分ごとに JSON へ保存し、`homehub-api` の
`/v1/server/disks` → Portal の `/api/disks` 経由で表示します。Portal は1分ごとに結果を取得します。
共有ディレクトリ内のファイル合計ではなく、共有先が載るファイルシステム全体の容量です。
収集結果が15分以上更新されなければ、その日時と更新遅延を表示します。
容量の取得失敗は機器データの表示に影響しません。

## ビルド・起動

Go は `.tool-versions`、Node.js は `apps/web/package.json` を参照してください。
依存のバージョンは勤怠管理アプリと揃え、`package-lock.json` に固定しています。

```sh
make build
./dist/home-portal serve
```

`http://127.0.0.1:8082/` を開きます。`make build` は初回または依存の変更時に
`npm ci --ignore-scripts` を実行し、Vue の型検査・ビルドと Go の埋め込みビルドを行います。
実行時に Node.js、Go、ソースディレクトリは不要です。

| 引数 | 既定値 | 用途 |
| --- | --- | --- |
| `--listen` | `127.0.0.1:8082` | Portal の待受 |
| `--homehub-url` | `http://127.0.0.1:8081` | Homehub API の接続先 |
| `--assets` | 未指定 | 開発時の外部 Vue ビルド結果 |

Homehub API への接続は Go サーバーから行います。ブラウザからループバックの API へ直接接続しません。
API が起動していなくても Portal とアプリのリンクは表示でき、機器データは取得失敗として表示します。
`/healthz` は Portal 自体の生存確認です。機器データの取得可否・鮮度を保証するものではありません。

`--homehub-url` は HTTP(S) のベース URL です。API のパスはサーバー側で固定しています。
接続・応答の期限は4秒、応答上限は1 MiBです。ブラウザの Cookie・Authorization は転送しません。

## 開発

初回は `make build` で画面を用意します。開発中は2つのターミナルで起動します。

```sh
# Go サーバー
go run ./cmd/home-portal serve --assets apps/web/dist
```

```sh
# Vue 開発サーバー。表示 URL は Vite が出力します。
cd apps/web
npm run dev
```

Vite は `/api` を Go の `127.0.0.1:8082` へ転送します。
ローカルから実機の API を確認する場合は、別途 `ssh -N -L 8081:127.0.0.1:8081 rha-hub` で転送できます。
アプリへのリンクは同一オリジンの `/kintai/`・`/cooking/` を前提とするため、開発サーバー単独ではリンク先アプリを配信しません。

## 検証

```sh
cd apps/web && npm ci --ignore-scripts
cd ../..
make test check integration
```

Go の race 検査・テスト、Vue のデータ処理テスト、gofmt・go vet・差分検査、
型検査・本番ビルド、空の作業ディレクトリから単一バイナリを起動する結合テストを実行します。
テストは一時 HTTP サーバーの架空データを使い、実機や実データには接続しません。
画面の表示・操作は別途ブラウザで確認します。

## Linux 向け配布

```sh
VERSION=0.1.0 make linux-amd64
```

`dist/home-portal-linux-amd64` を生成します。サービス定義の例は
[deploy/systemd/home-portal.service](deploy/systemd/home-portal.service) です。
サーバーへの配置と Caddy のルート追加は `home-server` 側で管理します。
既存の `/kintai/`・`/cooking/` の経路を維持し、Portal に割り当てる経路の残りを転送してください。
API は家庭内向けです。公開範囲と接続元の制限は配置先で設定します。

## 構成とデザイン

| パス | 内容 |
| --- | --- |
| `cmd/home-portal/` | 実行ファイルの入口 |
| `internal/command/` | CLI・起動・終了処理 |
| `internal/httpapi/` | Homehub API 連携・生存確認 |
| `internal/webui/` | Vue の埋め込み・外部アセット配信 |
| `apps/web/` | Vue SPA、固定機器一覧、画面のテスト |
| `integration/` | ビルド済みバイナリの結合テスト |

デザイン方針と共通スタイルの出典は [docs/design-guide.md](docs/design-guide.md) を参照してください。
