# docs サイト（Astro Starlight）

リポジトリ直下の `docs/`、課題シート（`drills/*/*/TASKS.md`）、`prompts/` を、そのまま読み込んでサイトにする。Markdown の正はそれらのファイルで、ここには設定とトップページ（`src/content/index.mdx`）だけを置く。ファイルとURLの対応は `src/lib/routes.mjs` にある。

```sh
npm ci           # 依存の取得（初回のみネットワークが必要）
npm run dev      # http://localhost:4321/offgrid/ で確認
npm run build    # dist/ に出力（dist/ はコミットしない）
```

- 公開：main に push すると、GitHub Actions（`.github/workflows/deploy-site.yml`）がビルドして GitHub Pages に出す
- 検索（Pagefind）と図（mermaid）は、どちらもビルドに同梱される。CDN は使わない
- 各ファイルは、先頭の `title:` がページの見出しになる。サイドバーは `astro.config.mjs` で、`docs/` の番号と `drills/` のディレクトリ名の順に組み立てる
- `docs/` は、サブディレクトリを作らず平らに保つ（サイドバーの自動生成と、リンクの書き換えが、この前提で作られている）
- `docs/` 内の相対リンク（`03-roadmap.md` など）は、`src/plugins/remark-doc-links.mjs` がサイト用のURLに書き換える

このサイトは AI が用意した補助ツールで、学習の課題ではない（T7 のビルド環境は、これを見ずに自分で組む）。
