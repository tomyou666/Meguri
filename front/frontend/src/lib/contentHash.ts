/** Markdown 本文をハッシュ入力用に正規化する（改行・前後空白）。 */
export function canonicalizeMarkdown(text: string): string {
	return text.replace(/\r\n/g, '\n').trim();
}

/** canonical markdown の SHA-256 十六進（node_results.content_hash と同一算法）。 */
export async function contentHashFromMarkdown(
	markdown: string,
): Promise<string> {
	const canonical = canonicalizeMarkdown(markdown);
	const data = new TextEncoder().encode(canonical);
	const buf = await crypto.subtle.digest('SHA-256', data);
	return Array.from(new Uint8Array(buf))
		.map((b) => b.toString(16).padStart(2, '0'))
		.join('');
}

/** links_json 比較用: URL 配列をソートして JSON 化。 */
export function canonicalizeLinksJson(
	links: string[] | null | undefined,
): string {
	if (!links?.length) return '[]';
	return JSON.stringify([...links].sort());
}
