import { defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';
import { docsSchema } from '@astrojs/starlight/schema';
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
};
