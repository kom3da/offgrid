// @ts-check
import { readdirSync } from 'node:fs';
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import mermaid from 'astro-mermaid';
import { remarkDocLinks } from './src/plugins/remark-doc-links.mjs';

const base = '/offgrid';
const repoUrl = 'https://github.com/kom3da/offgrid';

// One flat sidebar entry per file in ../docs, in file name order.
// (Autogeneration would nest everything under ".." > "docs" because the content lives outside src/.)
const pages = readdirSync(new URL('../docs', import.meta.url))
	.filter((file) => file.endsWith('.md'))
	.sort()
	.map((file) => file.replace(/\.md$/, ''));

// https://astro.build/config
export default defineConfig({
	site: 'https://kom3da.github.io',
	base,
	redirects: { '/': `${base}/00-overview/` },
	markdown: { remarkPlugins: [[remarkDocLinks, { base, repoUrl }]] },
	integrations: [
		// astro-mermaid must come before starlight. It bundles mermaid locally (no CDN).
		mermaid(),
		starlight({
			title: 'offgrid',
			description: 'AIなし・閉域でも「書ける・読める・直せる」力をゼロから付け直す学習カリキュラム',
			locales: { root: { label: '日本語', lang: 'ja' } },
			sidebar: pages,
			customCss: ['./src/styles/custom.css'],
			social: [{ icon: 'github', label: 'GitHub', href: repoUrl }],
		}),
	],
});
