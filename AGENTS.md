# home-portal

共通方針は親階層の AGENTS.md に従う。

- Go の単一バイナリが Vue SPA と Homehub API への読み取り API を配信する。
- 実行方法・設定は README.md、画面の方針は docs/design-guide.md を参照する。
- 機器名・MAC アドレス・アプリのリンクは apps/web/src/catalog.ts で管理する。観測値は homehub-api から取得し、架空の値を製品画面へ混ぜない。
- データベースや一覧編集の管理画面は設けない。取得失敗・未取得・古い観測を区別し、0 と null を保つ。
- Vue の構成とデザイントークンは kintai-kanri を基準にする。隣のリポジトリへの実行・ビルド依存を作らない。
- npm は apps/web/package-lock.json と .npmrc を使い、依存準備は npm ci --ignore-scripts とする。
- 本番ビルドは make build、Linux amd64 は VERSION=<version> make linux-amd64 とする。単純な go build は画面を埋め込まない。
- 変更に応じて make test、make check、make integration を使う。integration は Vue の型検査・ビルドと単一バイナリの起動を含む。通過済みの検証は変更がなければ繰り返さない。
- 画面変更ではブラウザの表示・操作も確認し、実行できない確認を明記する。API の検証には一時サーバーの架空データを使う。
- dist、node_modules、環境設定をコミットしない。サーバーへの配置とリバースプロキシは home-server 側で管理する。
