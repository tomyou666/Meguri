/** trim 後に空なら null */
export function normalizeToken(raw: string): string | null {
	const token = raw.trim();
	return token.length > 0 ? token : null;
}

/** 改行で分割。空行は捨てる。カンマはトークンの一部として残す */
export function splitTokens(raw: string): string[] {
	return raw
		.split(/[\n\r]+/)
		.map((s) => s.trim())
		.filter((s) => s.length > 0);
}

/** 重複は黙ってスキップ */
export function addToken(values: string[], raw: string): string[] {
	const token = normalizeToken(raw);
	if (!token || values.includes(token)) return values;
	return [...values, token];
}

/** splitTokens の各件を addToken で追加する */
export function addTokens(values: string[], raw: string): string[] {
	let next = values;
	for (const token of splitTokens(raw)) {
		next = addToken(next, token);
	}
	return next;
}

/** 1行1件の改行区切り文字列 */
export function tokensToCopyText(values: string[]): string {
	return values.join('\n');
}

export function removeTokenAt(values: string[], index: number): string[] {
	if (index < 0 || index >= values.length) return values;
	return values.filter((_, i) => i !== index);
}

export function removeLastToken(values: string[]): string[] {
	if (values.length === 0) return values;
	return values.slice(0, -1);
}
