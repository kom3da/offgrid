# F9：HTTPの基礎（課題シート）

この `README.md` は書き換えない。自分が書いたものは、同じディレクトリに別のファイルとして置く。

## ねらい

ブラウザとサーバーの間を流れているものを、自分の目で読めるようになる。Stage 2でAPIを作るときの土台になる。

## 平日に読むもの

『Webを支える技術』の、HTTPの基本、メソッド、ステータスコード、ヘッダを扱う章。

## 準備

```sh
mkdir -p ~/sandbox/f09 && cd ~/sandbox/f09
echo '<h1>hello</h1>' > index.html
```

シェルを2つ開き、片方で `python3 -m http.server 8000` を動かしておく。

## 課題

1. `curl -v http://localhost:8000/` を実行する。行の先頭の `*`、`>`、`<` は、それぞれ何を表すか
2. 出力の中から、リクエスト行、リクエストヘッダ、ステータス行、レスポンスヘッダ、ボディを見分ける。ヘッダとボディの境目は、どうなっているか
3. HTTPを手で話す：次を実行し、`curl` のときと同じ応答が返ることを確かめる。`\r\n` と、最後の空行は、何のためにあるか

   ```sh
   printf 'GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n' | nc localhost 8000
   ```

4. メソッド：`curl -I`（HEAD）、`curl -X POST -d 'a=1'`、`curl -X DELETE` を、自分のサーバーに送る。それぞれのステータスコードと、サーバー側のシェルに出るログを見比べる。存在しないパスを要求すると、どうなるか
5. ステータスコード：200、201、204、301、304、400、401、403、404、500、503 の意味を、表にまとめる。100の位は、何を表しているか
6. リダイレクト：`curl -I http://github.com` の、ステータスコードと `Location` ヘッダを読む。`-L` を付けると、何が変わるか
7. ヘッダ：`Content-Type`、`Content-Length`、`User-Agent`、`Host` の役割を調べる。`curl -H 'X-Test: 1' -A 'my-client/1.0'` を送り、`-v` の出力で、実際に送られたことを確かめる
8. HTTPS：`curl -v https://example.com` の出力のうち、HTTPが始まる前の部分（TLSのやり取り）を読む。証明書の発行者と有効期限は、どこに出ているか。証明書は、何を証明しているか
9. `curl -s -o /dev/null -w '%{http_code} %{time_total}\n' https://example.com` を実行し、`-s`、`-o`、`-w` がそれぞれ何をしているかを調べる
10. サーバー側のシェルに出たアクセスログを、F3のやり方で集計する（ステータスコードごとの件数）

## 完了条件の確かめ方

`curl -v https://example.com` の出力を `curl-v.txt` に保存する（`2>&1` が必要な理由も考える）。何も見ずに、`explain.md` に、出力の1行ずつの意味と、それがDNS・TCP・TLS・HTTPのどの段階かを書く。月曜に、AIに採点を頼む。

## 残すもの

`curl-v.txt`、`explain.md`、課題5の表、`notes.md`。
