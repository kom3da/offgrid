import { defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';
import { docsSchema } from '@astrojs/starlight/schema';

// The curriculum lives in ../docs (repository root) so that it stays readable on GitHub.
export const collections = {
	docs: defineCollection({
		loader: glob({ pattern: '**/*.md', base: '../docs' }),
		schema: docsSchema(),
	}),
};
