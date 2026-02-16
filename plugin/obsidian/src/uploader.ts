import { App, TFile, TFolder, normalizePath } from "obsidian";

export interface PluginSettings {
	baseUrl: string;
	defaultAuthor: string;
	defaultCategory: string;
	authToken: string;
}

export interface RawArticle {
	id: string | number;
	title: string;
}

export interface ParsedMeta {
	author?: string;
	category?: string;
	content: string;
}

export interface UploadTask {
	mdFile: TFile;
	title: string;
	mdContent: string;
	images: Array<{ name: string; data: ArrayBuffer }>;
}

const IMAGE_EXTENSIONS = new Set(["png", "jpg", "jpeg", "gif", "webp", "svg"]);

function joinUrl(baseUrl: string, path: string): string {
	const base = (baseUrl || "").replace(/\/+$/, "");
	const suffix = path.startsWith("/") ? path : `/${path}`;
	return `${base}${suffix}`;
}

function buildHeaders(token?: string, extra?: Record<string, string>): Record<string, string> {
	const headers: Record<string, string> = {
		Accept: "application/json",
		...(extra ?? {}),
	};
	if (token) {
		headers.Authorization = `Bearer ${token}`;
	}
	return headers;
}

function getNowString(): string {
	const pad = (n: number) => (n < 10 ? `0${n}` : String(n));
	const now = new Date();
	return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`;
}

export function extractMeta(md: string): ParsedMeta {
	let author: string | undefined;
	let category: string | undefined;
	let content = md;

	// 1) YAML frontmatter at top: --- ... ---
	if (md.startsWith("---")) {
		const end = md.indexOf("\n---", 3);
		if (end !== -1) {
			const fm = md.slice(3, end).split(/\r?\n/);
			for (const line of fm) {
				const m = line.match(/^\s*(author|category)\s*:\s*(.+)\s*$/i);
				if (!m) continue;
				const key = m[1].toLowerCase();
				const val = m[2].trim().replace(/^['"]|['"]$/g, "");
				if (!val) continue;
				if (key === "author") author = val;
				if (key === "category") category = val;
			}
			const after = md.slice(end + "\n---".length).replace(/^\r?\n/, "");
			content = after;
			return { author, category, content };
		}
	}

	// 2) fallback: top lines
	const lines = md.split(/\r?\n/);
	let i = 0;
	while (i < lines.length) {
		const line = lines[i];
		if (!line.trim()) {
			i++;
			continue;
		}
		const m = line.match(/^\s*(author|category)\s*:\s*(.+)\s*$/i);
		if (m) {
			const key = m[1].toLowerCase();
			const val = m[2].trim().replace(/^['"]|['"]$/g, "");
			if (key === "author" && !author && val) author = val;
			if (key === "category" && !category && val) category = val;
			i++;
			continue;
		}
		break;
	}
	if (i > 0) content = lines.slice(i).join("\n");

	return { author, category, content };
}

export async function fetchArticles(baseUrl: string, token?: string): Promise<RawArticle[]> {
	const url = joinUrl(baseUrl, "/api/allarticles");
	const res = await fetch(url, {
		method: "GET",
		headers: buildHeaders(token),
	});
	if (!res.ok) {
		const text = await res.text().catch(() => "");
		throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
	}

	const data = await res.json();
	let list: any[] = [];

	if (Array.isArray(data)) list = data;
	else if (Array.isArray(data?.articles)) list = data.articles;
	else if (Array.isArray(data?.data)) list = data.data;
	else if (Array.isArray(data?.data?.articles)) list = data.data.articles;
	else if (Array.isArray(data?.items)) list = data.items;
	else if (Array.isArray(data?.list)) list = data.list;

	return list.map((it, i) => ({
		id: it?.id ?? it?._id ?? it?.slug ?? it?.title ?? i,
		title: it?.title ?? String(it?.id ?? it?._id ?? i),
	}));
}

export async function deleteArticleByTitle(baseUrl: string, title: string, token?: string): Promise<void> {
	const url = joinUrl(baseUrl, `/api/article/${encodeURIComponent(title)}`);
	const res = await fetch(url, {
		method: "DELETE",
		headers: buildHeaders(token),
	});
	if (!res.ok) {
		const text = await res.text().catch(() => "");
		throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
	}
}

export async function collectImagesForMarkdown(app: App, mdFile: TFile): Promise<Array<{ name: string; data: ArrayBuffer }>> {
	const baseName = mdFile.basename;
	const parent = mdFile.parent?.path ?? "";
	const imageFolderPath = normalizePath(parent ? `${parent}/${baseName}` : baseName);

	const abstract = app.vault.getAbstractFileByPath(imageFolderPath);
	if (!(abstract instanceof TFolder)) return [];

	const images: Array<{ name: string; data: ArrayBuffer }> = [];
	for (const child of abstract.children) {
		if (!(child instanceof TFile)) continue;
		const ext = child.extension.toLowerCase();
		if (!IMAGE_EXTENSIONS.has(ext)) continue;
		const data = await app.vault.readBinary(child);
		images.push({ name: child.name, data });
	}
	return images;
}

export async function buildUploadTaskFromFile(app: App, mdFile: TFile): Promise<UploadTask> {
	if (mdFile.extension.toLowerCase() !== "md") {
		throw new Error("仅支持上传 Markdown 文件（.md）");
	}
	const mdContent = await app.vault.read(mdFile);
	const images = await collectImagesForMarkdown(app, mdFile);
	return {
		mdFile,
		title: mdFile.basename,
		mdContent,
		images,
	};
}

export async function buildUploadTasksFromFolder(app: App, folder: TFolder): Promise<UploadTask[]> {
	const tasks: UploadTask[] = [];

	const walk = async (dir: TFolder) => {
		for (const child of dir.children) {
			if (child instanceof TFolder) {
				await walk(child);
				continue;
			}
			if (!(child instanceof TFile)) continue;
			if (child.extension.toLowerCase() !== "md") continue;
			tasks.push(await buildUploadTaskFromFile(app, child));
		}
	};

	await walk(folder);
	return tasks;
}

export async function uploadArticle(
	baseUrl: string,
	token: string | undefined,
	title: string,
	mdContent: string,
	settings: PluginSettings
): Promise<void> {
	const meta = extractMeta(mdContent);
	const payload: Record<string, unknown> = {
		title,
		content: meta.content,
		update_at: getNowString(),
		author: meta.author || settings.defaultAuthor || "",
		category: meta.category || settings.defaultCategory || "General",
	};

	const url = joinUrl(baseUrl, `/api/article/${encodeURIComponent(title)}`);
	const res = await fetch(url, {
		method: "PUT",
		headers: buildHeaders(token, { "Content-Type": "application/json" }),
		body: JSON.stringify(payload),
	});
	if (!res.ok) {
		const text = await res.text().catch(() => "");
		throw new Error(`上传文章失败: HTTP ${res.status} ${res.statusText} ${text}`);
	}
}

export async function uploadImage(
	baseUrl: string,
	token: string | undefined,
	title: string,
	imageName: string,
	imageData: ArrayBuffer
): Promise<void> {
	const base64 = Buffer.from(imageData).toString("base64");
	const payload = {
		title,
		name: imageName,
		data: base64,
	};
	const url = joinUrl(baseUrl, "/api/image");
	const res = await fetch(url, {
		method: "PUT",
		headers: buildHeaders(token, { "Content-Type": "application/json" }),
		body: JSON.stringify(payload),
	});
	if (!res.ok) {
		const text = await res.text().catch(() => "");
		throw new Error(`上传图片失败: ${imageName} -> HTTP ${res.status} ${res.statusText} ${text}`);
	}
}

export async function uploadTask(task: UploadTask, settings: PluginSettings): Promise<void> {
	const baseUrl = (settings.baseUrl || "").trim();
	if (!baseUrl) throw new Error("未配置 baseUrl");

	await uploadArticle(baseUrl, settings.authToken, task.title, task.mdContent, settings);

	for (const img of task.images) {
		await uploadImage(baseUrl, settings.authToken, task.title, img.name, img.data);
	}
}

export async function uploadCurrentFileAsBlog(app: App, file: TFile, settings: PluginSettings): Promise<void> {
	const task = await buildUploadTaskFromFile(app, file);
	await uploadTask(task, settings);
}

export async function uploadFolderAsBlog(app: App, folder: TFolder, settings: PluginSettings): Promise<number> {
	const tasks = await buildUploadTasksFromFolder(app, folder);
	if (tasks.length === 0) return 0;

	for (const task of tasks) {
		await uploadTask(task, settings);
	}
	return tasks.length;
}
