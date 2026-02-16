export interface RawArticle {
	id: string | number;
	title: string;
}

function joinUrl(baseUrl: string, path: string): string {
	const base = (baseUrl || "").replace(/\/+$/, "");
	const suffix = path.startsWith("/") ? path : `/${path}`;
	return `${base}${suffix}`;
}

function authHeaders(token?: string): Record<string, string> {
	const headers: Record<string, string> = {
		Accept: "application/json",
	};
	if (token?.trim()) {
		headers.Authorization = `Bearer ${token.trim()}`;
	}
	return headers;
}

export function buildAllArticlesEndpoint(baseUrl: string): string {
	return joinUrl(baseUrl, "/api/allarticles");
}

export function buildArticleEndpoint(baseUrl: string, title: string): string {
	return joinUrl(baseUrl, `/api/article/${encodeURIComponent(title)}`);
}

export function buildImageEndpoint(baseUrl: string): string {
	return joinUrl(baseUrl, "/api/image");
}

export async function fetchArticles(baseUrl: string, token?: string): Promise<RawArticle[]> {
	const url = buildAllArticlesEndpoint(baseUrl);
	const res = await fetch(url, {
		method: "GET",
		headers: authHeaders(token),
	});

	if (!res.ok) {
		const text = await res.text().catch(() => "");
		throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
	}

	const data = await res.json();
	let list: any[] = [];

	if (Array.isArray(data)) {
		list = data;
	} else if (Array.isArray(data?.articles)) {
		list = data.articles;
	} else if (Array.isArray(data?.data)) {
		list = data.data;
	} else if (Array.isArray(data?.data?.articles)) {
		list = data.data.articles;
	} else if (Array.isArray(data?.items)) {
		list = data.items;
	} else if (Array.isArray(data?.list)) {
		list = data.list;
	}

	return list.map((it, i) => ({
		id: it?.id ?? it?._id ?? it?.slug ?? it?.title ?? i,
		title: it?.title ?? String(it?.id ?? it?._id ?? i),
	}));
}

export interface UploadArticlePayload {
	title: string;
	content: string;
	update_at: string;
	author?: string;
	category?: string;
}

export interface UploadImagePayload {
	title: string;
	name: string;
	data: string; // base64
}

export async function uploadArticle(
	baseUrl: string,
	payload: UploadArticlePayload,
	token?: string
): Promise<void> {
	const url = buildArticleEndpoint(baseUrl, payload.title);
	const headers = {
		...authHeaders(token),
		"Content-Type": "application/json",
	};

	const res = await fetch(url, {
		method: "PUT",
		headers,
		body: JSON.stringify(payload),
	});

	if (!res.ok) {
		const text = await res.text().catch(() => "");
		throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
	}
}

export async function uploadImage(
	baseUrl: string,
	payload: UploadImagePayload,
	token?: string
): Promise<void> {
	const url = buildImageEndpoint(baseUrl);
	const headers = {
		...authHeaders(token),
		"Content-Type": "application/json",
	};

	const res = await fetch(url, {
		method: "PUT",
		headers,
		body: JSON.stringify(payload),
	});

	if (!res.ok) {
		const text = await res.text().catch(() => "");
		throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
	}
}

export async function deleteArticleByTitle(
	baseUrl: string,
	title: string,
	token?: string
): Promise<void> {
	const url = buildArticleEndpoint(baseUrl, title);
	const res = await fetch(url, {
		method: "DELETE",
		headers: authHeaders(token),
	});

	if (!res.ok) {
		const text = await res.text().catch(() => "");
		throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
	}
}
