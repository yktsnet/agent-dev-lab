# agent-dev-lab

検証ループの待ちを測る Go の実験リポ。目的は `README.md` にある。
問い・完成条件・段取りは `PLAN.md` にある。

## コマンド

```bash
go vet ./...
go test ./...
go run ./cmd/lab seed-all   # MySQL を compose で起こし、config.toml の条件で全場面の投入を測る
go run ./cmd/lab down       # MySQL を止めてボリュームを消す
```

## 検証手段

上の `go vet` と `go test`。CI も同じ2つを回す。テストは MySQL を使わない。
投入や計測の変更は、場面の数を小さくした config をリポの外の一時ディレクトリに作って `-config` で流して確かめる。

## 構成の要点

- SQL は1回ごとに `docker compose exec` で mysql クライアントを起動して流す（`internal/dbx`）。この起動のコストが、測りたい「場面の数に比例する待ち」の正体なので、ドライバで接続を張りっぱなしにする形に変えない
- 場面は `config.toml` のシードから決定的に作る（`internal/scene`）。生成した場面をファイルとして置かない
- 生データは `results/raw/` の CSV に追記する。集計は生データから作り直せる形にし、生データを書き換えない

## 置かないもの

実在の業務ドメインの語・データ・接続情報を、コード・場面・文書に書かない。場面は中立のドメイン（イベントと予約）で作る。

## 作業の進め方

ブランチを切るときは worktree で切り、PR を出す。残っている作業は `issues/` に1件1ファイルで置き、済んだら `issues/done/` へ移す。
