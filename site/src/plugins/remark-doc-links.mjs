import { visit } from 'unist-util-visit';

// Rewrite relative Markdown links so they work on the built site:
//   01-roadmap.md            -> <base>/01-roadmap/
//   ../prompts/debug-drill.md -> the file on GitHub
export function remarkDocLinks({ base, repoUrl }) {
	return (tree) => {
		visit(tree, 'link', (node) => {
			const url = node.url;
			if (/^[a-z]+:|^#|^\//i.test(url)) return;
			const [path, hash = ''] = url.split('#');
			const anchor = hash ? `#${hash}` : '';
			if (path.startsWith('../')) {
				node.url = `${repoUrl}/blob/main/${path.slice(3)}${anchor}`;
				return;
			}
			if (!path.endsWith('.md')) return;
			node.url = `${base}/${path.replace(/^\.\//, '').replace(/\.md$/, '')}/${anchor}`;
		});
	};
}
