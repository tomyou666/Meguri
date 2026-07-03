/** trim 後に空なら null */
export function normalizeToken(raw: string): string | null {
	const token = raw.trim();
	return token.length > 0 ? token : null;
}

/** 重複は黙ってスキップ */
export function addToken(values: string[], raw: string): string[] {
	const token = normalizeToken(raw);
	if (!token || values.includes(token)) return values;
	return [...values, token];
}

export function removeTokenAt(values: string[], index: number): string[] {
	if (index < 0 || index >= values.length) return values;
	return values.filter((_, i) => i !== index);
}

export function removeLastToken(values: string[]): string[] {
	if (values.length === 0) return values;
	return values.slice(0, -1);
}
