---
title: "G16 contextと終了処理"
---

> トラック：[G：Go](../../../docs/11-track-g-go.md) ／ Stage 3 ／ 目安：メイン枠で2回 ／ 前提：G15、F4

## ねらい

「いつまで待つか」と「いつやめるか」を、プログラム全体に伝えられるようになる。応答しない相手をいつまでも待つプログラムや、止めるたびに処理中のリクエストを切り捨てるサーバーは、運用で必ず困る。Goでは、この2つを `context` という1つの仕組みで扱う。Ctrl+Cや `systemctl stop`（Stage 4）で止められたときに、やりかけの仕事を終えてから止まるサーバーを作る。

## このユニットで身につけること

- `select` で、複数のchannelのうち、先に準備ができたほうを処理する
- `context.WithTimeout` と `context.WithCancel` で、期限とキャンセルを作り、`cancel` を必ず呼ぶ
- `ctx.Done()` と `ctx.Err()` で、「なぜ止められたか」を見分ける
- `context` を、関数の引数としてHTTPリクエストまで伝える
- ハンドラの中で、クライアントが切断したことを `r.Context()` で知る
- `signal.NotifyContext` で、Ctrl+C（SIGINT）とSIGTERMを受け取る
- `http.Server` の `Shutdown` で、処理中のリクエストを終えてからサーバーを止める

## キーワード

- **`context.Context`**：期限、キャンセルの合図、リクエストに付随する値を、関数から関数へ運ぶための型。関数の最初の引数として渡すのが慣習
- **キャンセル**：「もう結果は要らないので、やめてよい」という合図。合図を受けた側が、自分で処理をやめる必要がある
- **デッドライン／タイムアウト**：「この時刻まで」／「今からこの時間だけ」という期限。過ぎると、自動でキャンセルされる
- **`select`**：複数のchannelの送受信を同時に待ち、準備ができたものを1つ実行する文
- **シグナル**：OSがプロセスに送る通知（F4）。Ctrl+CはSIGINT、`kill` の既定はSIGTERM。SIGKILLは、プログラム側では受け取れない
- **グレースフルシャットダウン**：新しい受け付けをやめ、処理中の仕事を終えてから止まること
- **goroutineリーク**：役目を終えたはずのgoroutineが、何かを待ったまま残り続けること。メモリや接続を少しずつ食いつぶす

## 平日に読むもの

- 『Go言語による並行処理』4章（Goでの並行処理パターン）
- 『初めてのGo言語』14章（コンテキスト）
- 『実用 Go言語』16章（エンタープライズなGoアプリケーションと並行処理）
- Go公式ブログの「Go Concurrency Patterns: Context」「Go Concurrency Patterns: Pipelines and cancellation」（オフライン用に保存しておく）
- `go doc context` の冒頭の説明、`go doc http.Server.Shutdown`、`go doc os/signal.NotifyContext`

## 準備

G14のAPIと、G15の死活監視を、このディレクトリにコピーして育てる。APIはこのディレクトリの直下、死活監視は `checker/` の下に置く（G15で `main` パッケージにしているなら、`go run ./checker` で動く）。

```sh
cp -r ../G14-middleware/{*.go,static,api.md} .
mkdir checker && cp ../G15-goroutines/*.go checker/
go mod init offgrid/g16-context-shutdown
go vet ./... && go test -race -cover ./...
```

実験用に、応答までに5秒かかるエンドポイント `GET /slow` を足しておく（認証は付けなくてよい）。

## 課題

1. `select` を試す。2つのgoroutineが、それぞれ別のchannelに、違う待ち時間の後で値を送る。`select` で、先に届いたほうだけを表示する。次に、片方を `time.After` に置き換え、「1秒以内に値が届かなければ、あきらめる」処理を書く。`default` を付けた `select` は、どう動きが変わるか
2. `context.WithTimeout` で、2秒の期限を持つ `ctx` を作る。「`ctx.Done()` が閉じるか、5秒たつか、早いほうで戻る」関数を書き、戻ったときの `ctx.Err()` を表示する。期限ではなく、`context.WithCancel` の `cancel` を1秒後に呼んで止めた場合、`ctx.Err()` はどう変わるか。2つのエラーを、`errors.Is`（G6）で見分ける
3. `context.WithTimeout` が返す `cancel` を、呼ばずに捨てるコード（`_` で受ける）を書き、`go vet` にかける。何と言われるか。期限が来れば自動で止まるのに、なぜ `cancel` を呼ぶ必要があるのかを、`go doc context.WithCancel` の説明から読み取る
4. 親の `ctx` から、子の `ctx` を作る（親は `WithCancel`、子は `WithTimeout`）。親をキャンセルすると、子はどうなるか。子をキャンセルすると、親はどうなるか。実験して、`notes.md` に図で書く
5. G15の「URLを1つ確認する関数」を、第1引数で `ctx` を受け取るように変える。`ctx` をHTTPリクエストに結び付ける方法を、`go doc net/http` の `NewRequest` で始まる関数の一覧から探す。3秒待たせる `httptest` の相手に対して、1秒の期限の `ctx` を渡すテストを書く。返ってくるエラーは、何を含んでいるか
6. ワーカープール全体を、1つの `ctx` で止められるようにする。300ミリ秒かかる相手を100個並べ、開始から1秒後にキャンセルする。確認が済んだ件数と、取りやめた件数を表示する。キャンセルの後、すべてのワーカーが終わったことを、`runtime.NumGoroutine()` の値（開始前と終了後）で確かめる。仕事のchannelに送る側は、キャンセルされたら、どうふるまうべきか
7. サーバー側で、クライアントの切断を知る。`GET /slow` を、「5秒たつか、`r.Context().Done()` が閉じるか、早いほうで戻る」ようにし、どちらで戻ったかをログに出す（G14の `slog`）。`curl localhost:8080/slow` を実行し、2秒後に `curl` をCtrl+Cで止める。サーバーのログには、何が出るか。ただの `time.Sleep(5 * time.Second)` に戻すと、どうなるか。切断を知ることで、サーバーは何を節約できるか
8. `signal.NotifyContext` を試す。まずは、HTTPとは関係のない小さなプログラムで、「1秒ごとに数を表示し続け、Ctrl+Cを受けたら `後片付け中` と表示して、1秒後に終わる」ものを書く。別のシェルから `kill`、`kill -TERM`、`kill -KILL` をそれぞれ送り、後片付けが動くかどうかを比べる（F4）。終了コード（`echo $?`）も、それぞれ記録する
9. **本題**：APIサーバーを、`http.ListenAndServe` から、`http.Server` の値を自分で作る形に変える。Ctrl+CかSIGTERMを受けたら `Shutdown` を呼び、処理中のリクエストを終えてから止まるようにする。`Shutdown` を待つ時間には、上限（例：10秒）を付ける。次を調べて、`notes.md` に書く
   - `Shutdown` を呼ぶと、`ListenAndServe` は何を返すか。それを、起動の失敗（例：ポートが使用中）と、どう見分けるか
   - `main` は、何を待ってから終わるべきか。`Shutdown` が終わる前に `main` が終わると、何が起きるか
10. 課題9の動きを確かめる。シェルAでサーバーを起動し、シェルBで `curl -i localhost:8080/slow` を実行し、5秒たつ前に、シェルAでCtrl+Cを押す。シェルBには、何が返るか。サーバーは、いつ終わるか。Ctrl+Cの後、`/slow` が終わるまでの間に、シェルCから新しく `curl localhost:8080/healthz` を送ると、どうなるか。変更前（`http.ListenAndServe` だけ）のコードでも同じ操作をし、結果を比べる
11. `/slow` の待ち時間を30秒に延ばし、`Shutdown` の上限は10秒のままで、課題10と同じ操作をする。何秒後に、何が起きるか。`Shutdown` の戻り値は、何か。上限を付けなかった場合、運用ではどんな困りごとが起きるか
12. `explain.md` に、何も見ずに書く：Ctrl+Cを押してからプロセスが終わるまでに、自分のサーバーの中で起きることを、順番に

### 発展（任意）

- G14のリクエストIDを、`context` に入れてハンドラまで運び、ハンドラの中のログにも同じIDを出す（`go doc context.WithValue`、`go doc http.Request.WithContext`）。キーの型について、`go doc context.WithValue` が勧めていることは何か。なぜか
- `http.Server` の `ReadTimeout`、`ReadHeaderTimeout`、`WriteTimeout`、`IdleTimeout` の意味を `go doc http.Server` で調べ、値を決めて設定する。`WriteTimeout` を3秒にすると、`/slow` はどうなるか
- 死活監視を、「10秒ごとに全URLを確認し続け、Ctrl+Cで、確認中のものを終えてから、集計を表示して止まる」常駐プログラムにする（`go doc time.Ticker`）

## 詰まりやすいところ

- Ctrl+Cを押すと、処理中のリクエストを待たずに、すぐ終わってしまう → `main` が、何を待ってから終わっているかを追う（課題9の2つ目の問い）
- Ctrl+Cを押しても、終わらなくなった → シグナルを受け取った後の処理が、実際に動いているかを、ログを足して確かめる。抜けられないときは、別のシェルから `kill -KILL`（F4）
- 正常に止めたのに、ログに `http: Server closed` がエラーとして出る、または終了コードが0にならない → `go doc http.ErrServerClosed` を読む
- 期限を付けたのに、HTTPリクエストが止まらない → 作った `ctx` が、リクエストまで渡っているかをたどる。途中で `context.Background()` を作り直していないか
- テストが、ときどき落ちる → 期限と、相手の待ち時間の差が小さすぎないか。「1秒の期限」と「1.1秒かかる相手」では、実行環境の揺れで結果が変わる
- `runtime.NumGoroutine()` が、終了後も開始前に戻らない → 戻らないgoroutineが、何を待っているのかを調べる。Ctrl+\（SIGQUIT）を送ると、Goのプログラムは、全goroutineのスタックトレースを出して終わる

## 完了条件の確かめ方

課題10の手順を実行する。(1) Ctrl+Cの後も、`curl -i localhost:8080/slow` に200とボディが最後まで返る。(2) その後でサーバーが終わり、`echo $?` が0になる。(3) 停止の途中に送った新しいリクエストは、受け付けられない。同じことを、Ctrl+Cの代わりに、別のシェルからの `kill -TERM <PID>` でも確かめる。`go vet ./... && go test -race -cover ./...` が通る。`explain.md` は、平日にAIに採点を頼む（[採点プロンプト](../../../prompts/grade-explain.md)）。

## 残すもの

- グレースフルシャットダウンに対応したAPIサーバー、`ctx` に対応した死活監視と、そのテスト
- `notes.md`（課題2〜4、7、8、10、11の観察結果。課題9の問いへの答え）
- `explain.md`

この `TASKS.md` はカリキュラムの一部なので、書き換えない。

## 次につながるユニット

- G17（DB連携）：`pgx` の関数は、どれも第1引数に `ctx` を取る。リクエストの `ctx` をDBの問い合わせまで渡し、終了時にはDBの接続も閉じる
- G19・G20（HTTPクライアント、SSE）：再試行の待ち時間を `ctx` で中断する。SSEでは、`r.Context()` でクライアントの切断を知る
- O2（systemd）：`systemctl stop` は、まずSIGTERMを送り、しばらく待ってからSIGKILLを送る。今回の終了処理が、そのまま効く
