import path from 'node:path';
import { visit } from 'unist-util-visit';
import { slugFor } from '../lib/routes.mjs';

// Rewrite relative Markdown links so they work on the built site.
// A link to a file that is a page on the site becomes a site URL; anything else in the
// repository (README.md, scripts/, templates/ ...) becomes a link to the file on GitHub.
export function remarkDocLinks({ base, repoUrl, repoRoot }) {
	return (tree, file) => {
		const current = path.relative(repoRoot, file.path ?? file.history?.[0] ?? '');
		visit(tree, 'link', (node) => {
			const url = node.url;
			if (/^[a-z][a-z0-9+.-]*:|^#|^\//i.test(url)) return;
			const [target, hash = ''] = url.split('#');
			const anchor = hash ? `#${hash}` : '';
			const repoPath = path.posix.normalize(path.posix.join(path.posix.dirname(current), target));
			if (repoPath.startsWith('..')) return;
			const slug = slugFor(repoPath);
			if (slug !== undefined) {
				node.url = slug === 'index' ? `${base}/${anchor}` : `${base}/${slug}/${anchor}`;
				return;
			}
			const kind = /\.[a-z0-9]+$/i.test(repoPath) ? 'blob' : 'tree';
			node.url = `${repoUrl}/${kind}/main/${repoPath}${anchor}`;
		});
	};
}
