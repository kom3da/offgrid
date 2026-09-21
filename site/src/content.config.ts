import { defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';
import { docsSchema, i18nSchema } from '@astrojs/starlight/schema';
import { i18nLoader } from '@astrojs/starlight/loaders';
import { contentPatterns, slugFor } from './lib/routes.mjs';

// The curriculum lives in the repository root (docs/, drills/**/TASKS.md, prompts/) so that it stays
// readable on GitHub and usable in a learner's copy. Only the top page lives inside site/.
export const collections = {
	docs: defineCollection({
		loader: glob({
			base: '..',
			pattern: contentPatterns,
			generateId: ({ entry }) => slugFor(entry) ?? entry,
		}),
		schema: docsSchema(),
	}),
	// 日本語だけのサイトだが、Starlight はこの入れ物を探しに来る。
	// 無いとビルドのたびに警告が出て、本物のエラーが埋もれる。
	i18n: defineCollection({ loader: i18nLoader(), schema: i18nSchema() }),
};
