// Maps a repository-relative Markdown path to its slug on the site.
// Used by the content loader, the link rewriter, and the sidebar, so the three never disagree.
//   docs/03-roadmap.md                        -> roadmap
//   drills/foundation/F03-pipeline/TASKS.md   -> units/f03-pipeline
//   prompts/grade-explain.md                  -> prompts/grade-explain
//   site/src/content/index.mdx                -> index (top page)
export function slugFor(repoPath) {
	let m;
	if (repoPath === 'site/src/content/index.mdx') return 'index';
	if ((m = repoPath.match(/^docs\/\d+-([^/]+)\.md$/))) return m[1];
	if ((m = repoPath.match(/^drills\/[^/]+\/([^/]+)\/TASKS\.md$/))) return `units/${m[1].toLowerCase()}`;
	if ((m = repoPath.match(/^prompts\/([^/]+)\.md$/))) return `prompts/${m[1]}`;
	return undefined;
}

export const contentPatterns = ['docs/*.md', 'drills/*/*/TASKS.md', 'prompts/*.md', 'site/src/content/index.mdx'];
