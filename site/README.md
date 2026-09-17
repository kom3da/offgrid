# docs サイト（Astro Starlight）

リポジトリ直下の `docs/` を、そのまま読み込んでサイトにする。Markdown の正は `docs/` で、ここには設定だけを置く。

```sh
npm ci           # 依存の取得（初回のみネットワークが必要）
npm run dev      # http://localhost:4321/offgrid/ で確認
npm run build    # dist/ に出力（dist/ はコミットしない）
```

- 公開：main に push すると、GitHub Actions（`.github/workflows/deploy-site.yml`）がビルドして GitHub Pages に出す
- 検索（Pagefind）と図（mermaid）は、どちらもビルドに同梱される。CDN は使わない
- `docs/` の各ファイルは、先頭の `title:` がページの見出しになる
- `docs/` 内の相対リンク（`01-roadmap.md` など）は、`src/plugins/remark-doc-links.mjs` がサイト用のURLに書き換える

このサイトは AI が用意した補助ツールで、学習の課題ではない（T7 のビルド環境は、これを見ずに自分で組む）。
