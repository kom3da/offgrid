---
title: "G20 Server-Sent Events"
---

> トラック：[G：Go](../../../docs/11-track-g-go.md) ／ Stage 3 ／ 目安：メイン枠で2回 ／ 前提：G16、G14

## ねらい

これまでのHTTPは「1つ要求して、1つ返す」だった。今度は、サーバーからクライアントへ、つないだまま値を流し続ける。仕組みはSSE（Server-Sent Events）。特別なライブラリは要らず、決まった形のテキストを書いて、こまめに送り出すだけで、ブラウザの `EventSource` が受け取れる。難しいのは「流す」ことより「やめる」こと。クライアントが切断したのに気づかず、goroutineが残り続けるサーバーは、日ごとに重くなっていく。切断を `context` で検知して、後片付けまで含めて正しく止まるエンドポイントを作る。

## このユニットで身につけること

- SSEの書式（`data:`、`event:`、`id:`、空行）を、仕様書から読み取って書く
- `http.Flusher` で、レスポンスの一部をすぐに送り出す
- 1秒ごとに値を送り続けるエンドポイントを作り、`curl -N` で確かめる
- クライアントの切断を `r.Context()` で検知し、値を作るgoroutineを止める
- `runtime.NumGoroutine()` と `pprof` で、goroutineが残っていないことを確かめる
- G14のミドルウェアとG16の終了処理が、流れ続けるレスポンスと両立するかを確かめる
- `httptest` で、SSEのエンドポイントと、切断後の後片付けをテストする

## キーワード

- **SSE（Server-Sent Events）**：1つのHTTPレスポンスを閉じずに、サーバーからイベントを送り続ける仕組み。向きは、サーバーからクライアントの一方向
- **`text/event-stream`**：SSEのレスポンスの `Content-Type`
- **イベント**：SSEで送る1つのまとまり。`data:` で始まる行と、終わりを表す空行からなる
- **バッファリング**：書き込んだデータを、すぐには送らず、ためてからまとめて送ること。効率はよいが、「すぐ届けたい」ときにはじゃまになる
- **フラッシュ**：ためているデータを、今すぐ送り出すこと。Goでは `http.Flusher`
- **`EventSource`**：ブラウザが持つ、SSEを受け取るための仕組み（T11で使う）。自動で再接続する
- **長時間接続**：何分も何時間も閉じない接続。タイムアウトや終了処理の前提が、ふつうのAPIと変わる
- **goroutineリーク**：役目を終えたはずのgoroutineが残り続けること（G16）。長時間接続では、接続の数だけ起きうる
- **pprof**：動いているプログラムのgoroutineやメモリの状態を見るための、標準ライブラリの道具

## 平日に読むもの

- HTML Standard（WHATWGの仕様書。オフライン用に保存しておく）の「Server-sent events」の節。とくに、イベントストリームの書式と、各フィールドの意味
- MDN（DevDocsのオフライン版）の「Using server-sent events」と「EventSource」。`EventSource` がリクエストに何を付けられて、何を付けられないか
- 『実用 Go言語』11章（HTTPサーバー）、13章（ログとオブザーバビリティ）
- 『初めてのGo言語』14章（コンテキスト）
- `go doc http.Flusher`、`go doc http.ResponseController`、`go doc http.Request.Context`（いつキャンセルされるかの説明）、`go doc runtime.NumGoroutine`、`go doc net/http/pprof`、`man curl`（`-N`）

## 準備

G16のAPI（グレースフルシャットダウンとミドルウェアが付いたもの）を土台にする。構成はG18に合わせる。

```sh
go mod init offgrid/g20-sse
mkdir -p cmd/api internal
cp ../G16-context-shutdown/*.go cmd/api/ && cp -r ../G16-context-shutdown/static ../G16-context-shutdown/api.md .
go vet ./... && go test -race -cover ./...
```

SSEの処理は `internal/` の下の新しいパッケージに書き、`cmd/api` から `GET /events` として登録する。認証（G14）は、最初は外しておき、課題9で付ける。

## 課題

1. 仕様書を読んで、`notes.md` に書き出す：イベントストリームの1つのイベントは、どの行で始まり、何で終わるか。`data:`、`event:`、`id:`、`retry:` のそれぞれの意味。`:` で始まる行の扱い。`data:` が2行続いたとき、クライアントは何を受け取るか。レスポンスに付けるべき `Content-Type`
2. `GET /events` を作る。まずは、`Content-Type` を付け、`data:` の行と空行を1秒おきに5回書いて終わるだけにする。フラッシュはしない。`curl -N localhost:8080/events` で、行はいつ届くか（1秒おきか、5秒後にまとめてか）。`-N` を外すと、どう変わるか。`man curl` で `-N` の意味を読む
3. `http.Flusher` を使う。`w` を `http.Flusher` に型アサーション（G8）し、イベントを書くたびに `Flush` する。課題2の観察を繰り返す。型アサーションが失敗する（`http.Flusher` を満たさない）のは、どんなときか。失敗したら、ハンドラは何を返すべきか
4. 書式を試す：`event: tick` を付けたイベント、`data:` が2行あるイベント、`: ping` というコメント行、`id:` を付けたイベントを、順に送る。`curl -N` には、それぞれどう表示されるか（`curl` は解釈せず、そのまま表示する）。空行を抜いた「壊れたイベント」を送ると、`curl` では見分けられるか。ブラウザで見分けるのは、T11でやる
5. 終わりのないストリームにする。`time.Ticker`（G16の発展）で1秒ごとに、現在時刻か連番を送り続ける。各イベントを送るたびに `slog` にログを出す。`curl -N` を数秒で Ctrl+C した後、サーバーのログはどうなるか。ハンドラのループは、止まるか、続くか。`Write` の戻り値のエラーを見ていれば、何回目の書き込みで気づけるか
6. 切断を検知する。ループを `select` にして、ティッカーと `r.Context().Done()` の両方を待つ（G16の課題7）。`curl` を止めてから、ログに「切断」が出るまで、何秒かかるか。`go doc http.Request.Context` を読み、`r.Context()` がキャンセルされる3つの場合を書く。ハンドラが `return` した後、この `ctx` はどうなるか
7. 値を作る側を、ハンドラとは別のgoroutineにする（例：`internal/` に「1秒ごとに値をchannelに送る」関数を置き、ハンドラはchannelから受け取って書く。G15）。切断されたとき、このgoroutineは止まるか。止まるように、`ctx` を渡し、ティッカーも止める。goroutineの中で「channelに送る」と「`ctx.Done()`」を、どう同時に待つか
8. **本題**：goroutineが残らないことを確かめる。`runtime.NumGoroutine()` の値を返すエンドポイント（`GET /debug/goroutines`）を足す。手順：(1) 何もつないでいないときの値を記録する。(2) `curl -N` を5つ、バックグラウンド（F4の `&`）で起動し、値を見る。(3) 5つとも `kill` し、数秒待って、値を見る。(1) と (3) は同じか。同じでないなら、`go doc net/http/pprof` を読んで `pprof` のハンドラを登録し、`/debug/pprof/goroutine?debug=2` で、残っているgoroutineがどの行で何を待っているかを見て、直す。課題7の前のコード（goroutineを止めていない版）でも同じ手順を行い、値の違いを記録する
9. ミドルウェアと終了処理との両立：(a) G14のアクセスログのミドルウェアの内側に `/events` を置くと、課題3の型アサーションは通るか。通らないなら、G14で作った `http.ResponseWriter` の包みに何が足りないか。`http.ResponseController` を使う書き方も `go doc` で調べる。アクセスログの1行は、いつ書かれるか。(b) `http.Server` に `WriteTimeout`（G16の発展）を10秒で設定すると、ストリームはどうなるか。(c) `curl -N` をつないだまま、サーバーにCtrl+Cを送る。`Shutdown` は、上限の時間まで待つか、すぐ戻るか。流している最中の接続を、`Shutdown` に合わせて閉じるには、何が要るか（`go doc http.Server.RegisterOnShutdown`、`go doc http.Server.BaseContext` の一覧から考える）
10. 認証：`/events` を、G14の認証ミドルウェアの内側に置く。`curl -N -H 'Authorization: Bearer ...'` でつながり、ヘッダなしでは401になることを確かめる。MDNの `EventSource` のページで、ブラウザからこのヘッダを付けられるかを調べ、結果を `notes.md` に書く。付けられないなら、代わりの方法の候補を思いつくだけ挙げておく（比べるのはT11、決めるのはC1・C4）
11. `httptest` でテストを書く。(a) `httptest.NewRecorder` は `http.Flusher` を満たすか（`go doc httptest.ResponseRecorder`）。期限付きの `ctx` を持つリクエスト（`http.Request.WithContext`）をハンドラに渡し、`Content-Type` と、最初のイベントの書式（`data:` の行と空行）を確かめる。(b) `httptest.NewServer` で本物のサーバーを立て、`ctx` 付きのクライアントで3イベント読んでから `cancel` する。その後、`runtime.NumGoroutine()` が開始前の値に戻ることを、少し待ちながら確かめる。`go test -race -count=5` で安定して通ること
12. `explain.md` に、何も見ずに書く：`curl -N` がつないでから Ctrl+C するまでに、サーバーの中で何が起きるか（ハンドラ、フラッシュ、値を作るgoroutine、`ctx`、後片付け）。goroutineが残る書き方と、残らない書き方の違い

### 発展（任意）

- `id:` を付けて送り、クライアントが `Last-Event-ID` ヘッダ付きで再接続してきたら、その続きから送る。「続き」を覚えておくのは誰か
- 複数のクライアントに、同じ値を同時に配る（1つの値を作るgoroutineから、接続ごとのchannelへ配る。G15の `Mutex` で接続の一覧を守る）。切断したクライアントのchannelを、一覧から確実に外す
- 15秒ごとにコメント行（`: ping`）を送り、途中の機器に接続を切られにくくする。なぜコメント行でよいのか
- `retry:` フィールドで、`EventSource` の再接続の間隔を指定する。T11で効いているかを確かめる

## 詰まりやすいところ

- `curl -N` に何も届かず、5秒後にまとめて届く → `Flush` を呼んでいるか。呼んでいるなら、`w` を包んでいるミドルウェアが `http.Flusher` を満たしているか（課題9）
- 型アサーションでpanicする → 2つの戻り値を受ける形（G8）にしているか
- `curl` を止めてもログが続く、または `broken pipe` が出続ける → `select` に `r.Context().Done()` が入っているか。値を作るgoroutineに `ctx` が渡っているか
- `NumGoroutine` が開始前に戻らない → `/debug/pprof/goroutine?debug=2` で、残っているgoroutineの行を見る。channelへの送信で止まっていないか。ティッカーを `Stop` しているか
- テストがときどき落ちる → 切断からgoroutineが終わるまでには、少し時間がかかる。「すぐ確かめる」のではなく、上限付きで待ちながら確かめているか
- 10秒でストリームが切れる → `http.Server` の `WriteTimeout` を疑う（課題9 (b)）

## 完了条件の確かめ方

```sh
go vet ./... && go test -race -count=5 -cover ./...
go run ./cmd/api &   # 別のシェルで
curl -N -H "Authorization: Bearer $API_TOKEN" localhost:8080/events
```

- `curl -N` で、1秒ごとに値が届き続ける（60秒以上）。止めると、サーバーのログに切断が記録され、値の送信が止まる
- 課題8の手順で、5つの接続を切った後の `NumGoroutine` が、つなぐ前の値に戻る
- テストに、書式の確認と、切断後にgoroutineが戻ることの確認が含まれ、`-race -count=5` で通る
- `explain.md` を、何も見ずに書いてある。平日に、AIに採点を頼む（[`prompts/grade-explain.md`](../../../prompts/grade-explain.md)）

## 残すもの

- `cmd/api/`、`internal/` のSSEのコードとテスト、更新した `api.md`（`/events` の書式と、認証の付け方）
- `notes.md`（課題1・2・4〜6・8〜10の観察結果と、課題10の候補）、`explain.md`
- `API_TOKEN` の値と `.env` は、コミットしない

この `TASKS.md` はカリキュラムの一部なので、書き換えない。

## 次につながるユニット

- T11（SSEの受信）：ブラウザの `EventSource` で、この `/events` を受けてReactで表示する。`EventSource` はリクエストにヘッダを付けられないので、課題10の候補を、そこで2つに絞って比べる
- C4（卒業制作の配信）：メトリクスをSSEで配る。SSEの認証をどうするか（クエリ文字列、Cookie、別の方法）を、C1でADRとして決めてから実装する。課題9 (c) の終了処理も、そのまま要る
- C7（障害試験）：server停止からの復帰で、ブラウザ側が自動で再接続することを確かめる試験に、課題8の手順（接続数とgoroutine数）を組み込む
