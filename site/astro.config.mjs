// @ts-check
import { readdirSync, existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'astro/config';
import { unified } from '@astrojs/markdown-remark';
import starlight from '@astrojs/starlight';
import mermaid from 'astro-mermaid';
import starlightThemeRapide from 'starlight-theme-rapide';
import { remarkDocLinks } from './src/plugins/remark-doc-links.mjs';
import { slugFor } from './src/lib/routes.mjs';

const base = '/offgrid';
const repoUrl = 'https://github.com/kom3da/offgrid';
const repoRoot = fileURLToPath(new URL('..', import.meta.url));

const doc = (file) => slugFor(`docs/${file}.md`);

// Unit pages of one track, in directory-name order (F01, F02, ...), read from drills/<dir>/.
function units(dir) {
	const root = new URL(`../drills/${dir}/`, import.meta.url);
	if (!existsSync(root)) return [];
	return readdirSync(root)
		.filter((name) => existsSync(new URL(`${name}/TASKS.md`, root)))
		.sort()
		.map((name) => slugFor(`drills/${dir}/${name}/TASKS.md`));
}

const track = (label, file, dir) => ({ label, collapsed: true, items: [doc(file), ...units(dir)] });

const prompts = readdirSync(new URL('../prompts/', import.meta.url))
	.filter((name) => name.endsWith('.md'))
	.sort()
	.map((name) => slugFor(`prompts/${name}`));

// https://astro.build/config
export default defineConfig({
	site: 'https://kom3da.github.io',
	base,
	// Astro 7 defaults to a new Markdown processor; remark plugins need the unified one.
	markdown: { processor: unified({ remarkPlugins: [[remarkDocLinks, { base, repoUrl, repoRoot }]] }) },
	integrations: [
		// astro-mermaid must come before starlight. It bundles mermaid locally (no CDN).
		mermaid(),
		starlight({
			title: 'offgrid',
			description: 'AIなし・閉域でも「書ける・読める・直せる」力をゼロから付け直す学習カリキュラム',
			locales: { root: { label: '日本語', lang: 'ja' } },
			social: [{ icon: 'github', label: 'GitHub', href: repoUrl }],
			plugins: [starlightThemeRapide()],
			customCss: ['./src/styles/custom.css'],
			sidebar: [
				{ label: 'はじめに', items: [doc('01-overview'), doc('02-getting-started')] },
				{
					label: '進め方',
					items: [doc('03-roadmap'), doc('04-ai-free-day'), doc('05-debug-and-reading'), doc('06-evaluation')],
				},
				{
					label: 'カリキュラム',
					items: [
						track('F：基礎', '10-track-f-foundation', 'foundation'),
						track('G：Go', '11-track-g-go', 'go'),
						track('T：Webフロントエンド', '12-track-t-web', 'ts'),
						track('D：PostgreSQL', '13-track-d-postgresql', 'sql'),
						track('O：運用と閉域', '14-track-o-ops', 'infra'),
						track('C：卒業制作', '15-track-c-capstone', 'capstone'),
					],
				},
				{
					label: '資料',
					items: [doc('20-resources'), { label: 'AIへの頼み方', collapsed: true, items: prompts }],
				},
				{ label: '管理者向け', items: [doc('30-revision')] },
			],
		}),
	],
});
