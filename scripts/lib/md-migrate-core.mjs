/**
 * Shared Markdown → SQL migration helpers (Hexo-style frontmatter; hashes match server domain).
 */

import { createHash, randomBytes } from "node:crypto";
import { readdir } from "node:fs/promises";
import { join, relative, resolve } from "node:path";
import { DateTime } from "luxon";
import YAML from "yaml";

/** @param {string} s */
export function dollarQuote(s) {
	for (let k = 0; k < 50; k++) {
		const tag = "m" + randomBytes(16).toString("hex");
		const delim = `$${tag}$`;
		if (!s.includes(delim)) return `${delim}${s}${delim}`;
	}
	throw new Error("cannot delimit string for SQL");
}

/** @param {string} pathStem @param {string} title */
export function fallbackShortUrlArticle(pathStem, title) {
	let raw = pathStem.replace(/[^a-zA-Z0-9-]+/g, "-");
	raw = raw.replace(/-+/g, "-").replace(/^-|-$/g, "").toLowerCase();
	if (raw.length >= 2) return raw.slice(0, 255);
	const h = createHash("md5")
		.update(`${title}\0${pathStem}`, "utf8")
		.digest("hex")
		.slice(0, 16);
	return `import-${h}`;
}

/** @param {string} pathStem @param {string} title */
export function fallbackShortUrlMoment(pathStem, title) {
	let raw = pathStem.replace(/[^a-zA-Z0-9-]+/g, "-");
	raw = raw.replace(/-+/g, "-").replace(/^-|-$/g, "").toLowerCase();
	if (raw.length >= 2) return raw.slice(0, 255);
	const h = createHash("md5")
		.update(`${title}\0${pathStem}`, "utf8")
		.digest("hex")
		.slice(0, 16);
	return `moment-${h}`;
}

/** @param {string} isoTs */
export function timestamptzLiteral(isoTs) {
	return `${dollarQuote(isoTs)}::timestamptz`;
}

/**
 * @param {unknown} value
 * @param {string} defaultTz e.g. Asia/Shanghai
 */
export function toRfc3339(value, defaultTz) {
	if (value instanceof Date) {
		const dt = DateTime.fromJSDate(value, { zone: "utc" });
		if (!dt.isValid) throw new Error(`invalid Date: ${value}`);
		return dt.setZone(defaultTz).toISO({ suppressMilliseconds: true }) ?? dt.toISO();
	}
	if (typeof value === "string") {
		let s = value.trim();
		for (const fmt of ["yyyy-MM-dd HH:mm:ss", "yyyy-MM-dd HH:mm", "yyyy-MM-dd"]) {
			const dt = DateTime.fromFormat(s, fmt, { zone: defaultTz });
			if (dt.isValid) return dt.toISO({ suppressMilliseconds: true }) ?? dt.toISO();
		}
		if (s.endsWith("Z")) s = s.slice(0, -1) + "+00:00";
		let dt = DateTime.fromISO(s, { setZone: true });
		if (!dt.isValid) throw new Error(`unsupported date value: ${JSON.stringify(value)}`);
		if (dt.offset === 0 && !/[zZ]|[+-]\d{2}:?\d{2}$/.test(value.trim())) {
			dt = DateTime.fromISO(s, { zone: defaultTz });
		}
		return dt.toISO({ suppressMilliseconds: true }) ?? dt.toISO();
	}
	if (value && typeof value === "object" && "toISOString" in value && typeof value.toISOString === "function") {
		return toRfc3339(new Date(/** @type {any} */ (value).toISOString()), defaultTz);
	}
	throw new Error(`unsupported date value: ${JSON.stringify(value)}`);
}

/** @param {string} text */
export function splitFrontmatter(text) {
	if (!text.startsWith("---")) return [{}, text];
	const re = /^---[ \t]*\r?\n([\s\S]*?)\r?\n---[ \t]*\r?\n([\s\S]*)$/;
	const m = text.match(re);
	if (!m) return [{}, text];
	const rawYaml = m[1];
	let body = m[2];
	if (body.startsWith("\n")) body = body.slice(1);
	let meta = YAML.parse(rawYaml);
	if (meta == null) meta = {};
	if (typeof meta !== "object" || Array.isArray(meta)) {
		throw new Error("frontmatter YAML must be an object");
	}
	return [meta, body];
}

/** @param {string} body */
export function firstHeadingTitle(body) {
	for (const line of body.split(/\r?\n/)) {
		const t = line.trim();
		if (t.startsWith("#")) {
			const title = t.replace(/^#+\s*/, "").trim();
			return title || null;
		}
	}
	return null;
}

/** @param {string} permalink */
export function permalinkToSlug(permalink) {
	const p = String(permalink).trim().replace(/^["']|["']$/g, "").replace(/^\/+|\/+$/g, "");
	if (!p) return null;
	const parts = p.split("/");
	return parts[parts.length - 1] || null;
}

/**
 * @param {Record<string, unknown>} fm
 * @param {string} body
 * @param {string} fileStem
 * @param {string} defaultTz
 */
export function frontmatterToWork(fm, body, fileStem, defaultTz) {
	/** @type {Record<string, unknown>} */
	const out = { content: body };

	const titleRaw = fm.title;
	if (titleRaw != null && String(titleRaw).trim()) {
		out.title = String(titleRaw).trim();
	} else {
		const h = firstHeadingTitle(body);
		out.title = h ?? fileStem;
	}

	if (fm.cover) out.cover = String(fm.cover).trim();

	let slug = null;
	for (const key of ["abbrlink", "slug"]) {
		if (fm[key] != null && String(fm[key]).trim()) {
			slug = String(fm[key]).trim();
			break;
		}
	}
	if (slug == null && fm.permalink) slug = permalinkToSlug(String(fm.permalink));
	if (slug) out.shortUrl = slug;

	const d = fm.date ?? fm.createdAt ?? fm.created ?? fm.updated;
	if (d == null) {
		throw new Error(
			"frontmatter must include one of: date, createdAt, created, updated (for content time)"
		);
	}
	out.createdAt = toRfc3339(d, defaultTz);

	let tags = fm.tags;
	if (tags == null) tags = fm.tag;
	/** @type {string[]} */
	const names = [];
	if (Array.isArray(tags)) {
		for (const t of tags) {
			const s = String(t).trim();
			if (s) names.push(s);
		}
	} else if (typeof tags === "string" && tags.trim()) {
		names.push(tags.trim());
	}
	if (names.length) out._tag_names = names;

	if (fm.summary != null) out.summary = String(fm.summary);
	else out.summary = "";

	return out;
}

/**
 * @param {Record<string, unknown>} work mutable
 * @param {Record<string, number>} tagsMap
 */
export function applyTagsMap(work, tagsMap) {
	const names = /** @type {string[] | undefined} */ (work._tag_names);
	delete work._tag_names;
	if (!names?.length || !Object.keys(tagsMap).length) return [];
	/** @type {number[]} */
	const ids = [];
	/** @type {string[]} */
	const missing = [];
	for (const n of names) {
		if (n in tagsMap) {
			const tid = tagsMap[n];
			if (!ids.includes(tid)) ids.push(tid);
		} else missing.push(n);
	}
	if (ids.length) work.tagIds = ids;
	return missing;
}

/**
 * @param {Record<string, unknown>} work
 * @param {Record<string, number>} tagsMap
 * @param {number | null} categoryId
 */
export function workToApiPayload(work, tagsMap, categoryId) {
	const w = { ...work };
	const missing = applyTagsMap(w, tagsMap);

	/** @type {Record<string, unknown>} */
	const payload = {
		title: w.title,
		content: w.content,
		summary: w.summary ?? "",
		isPublished: Boolean(w.isPublished ?? true),
		isTop: Boolean(w.isTop ?? false),
		isOriginal: Boolean(w.isOriginal ?? true),
		createdAt: w.createdAt,
	};
	if (w.shortUrl) payload.shortUrl = String(w.shortUrl).trim();
	if (categoryId != null) payload.categoryId = categoryId;
	else if (w.categoryId != null) payload.categoryId = w.categoryId;
	if (w.tagIds != null) payload.tagIds = w.tagIds;
	if ("allowComment" in w) payload.allowComment = w.allowComment;
	if (w.cover) payload.cover = w.cover;
	if (w.leadIn != null) payload.leadIn = w.leadIn;
	if (w.aiSummary != null) payload.aiSummary = w.aiSummary;
	if (w.extInfo != null) payload.extInfo = w.extInfo;
	if (w.views != null) payload.views = w.views;
	return { payload, missingTags: missing };
}

/** @param {string} title @param {string | null | undefined} leadIn @param {string} content */
export function articleContentHash(title, leadIn, content) {
	const h = createHash("md5");
	h.update(title, "utf8");
	h.update(Buffer.from([0]));
	h.update(leadIn ?? "", "utf8");
	h.update(Buffer.from([0]));
	h.update(content, "utf8");
	return h.digest("hex");
}

/** @param {string} summary @param {string} body @param {number} [limit] */
export function buildSummary(summary, body, limit = 200) {
	if (summary && summary.trim()) return summary.trim();
	const chars = [...body];
	if (chars.length <= limit) return body;
	return chars.slice(0, limit).join("");
}

/** @param {string} title @param {string} summary @param {string} content */
export function momentContentHash(title, summary, content) {
	const h = createHash("md5");
	h.update(title, "utf8");
	h.update(Buffer.from([0]));
	h.update(summary, "utf8");
	h.update(Buffer.from([0]));
	h.update(content, "utf8");
	return h.digest("hex");
}

/**
 * @param {Record<string, unknown>} fm
 * @param {string} body
 * @param {string} fileStem
 * @param {string} defaultTz
 * @param {Record<string, number>} topicsMap
 * @param {number | null} columnId
 */
export function frontmatterToMomentPayload(fm, body, fileStem, defaultTz, topicsMap, columnId) {
	const work = frontmatterToWork(fm, body, fileStem, defaultTz);
	const title = String(work.title);
	const summary = buildSummary(String(work.summary ?? ""), body);
	const content = String(work.content);
	const shortUrl = work.shortUrl ? String(work.shortUrl) : fallbackShortUrlMoment(fileStem, title);

	let img = null;
	if (fm.cover) img = String(fm.cover).trim();
	else if (fm.img) img = String(fm.img).trim();

	/** @type {Record<string, unknown>} */
	const payload = {
		title,
		summary,
		content,
		shortUrl,
		createdAt: work.createdAt,
		isPublished: Boolean(work.isPublished ?? true),
		isTop: Boolean(work.isTop ?? false),
		isOriginal: Boolean(work.isOriginal ?? true),
		views: Number(work.views ?? 0) || 0,
		img,
	};
	if (columnId != null) payload.columnId = columnId;

	let names = /** @type {string[] | undefined} */ (work._tag_names);
	if (!names) {
		names = [];
		let tags = fm.tags;
		if (tags == null) tags = fm.tag;
		if (Array.isArray(tags)) {
			for (const t of tags) {
				const s = String(t).trim();
				if (s) names.push(s);
			}
		} else if (typeof tags === "string" && tags.trim()) {
			names.push(tags.trim());
		}
	}

	/** @type {number[]} */
	const ids = [];
	/** @type {string[]} */
	const missing = [];
	for (const n of names) {
		if (n in topicsMap) {
			const tid = topicsMap[n];
			if (!ids.includes(tid)) ids.push(tid);
		} else missing.push(n);
	}
	payload.topicIds = ids;
	return { payload, missingTopics: missing };
}

/** @param {string} path */
export async function loadTagsMap(path) {
	const { readFile } = await import("node:fs/promises");
	const data = JSON.parse(await readFile(path, "utf8"));
	if (data === null || typeof data !== "object" || Array.isArray(data)) {
		throw new Error('tags map must be a JSON object: {"tag name": id, ...}');
	}
	/** @type {Record<string, number>} */
	const out = {};
	for (const [k, v] of Object.entries(data)) {
		out[String(k)] = Number(v);
	}
	return out;
}

/**
 * @param {string} dir absolute or cwd-relative
 * @returns {AsyncGenerator<string>}
 */
export async function* walkMarkdownFiles(dir) {
	const abs = resolve(dir);
	const entries = await readdir(abs, { withFileTypes: true });
	for (const e of entries) {
		const p = join(abs, e.name);
		if (e.isDirectory()) yield* walkMarkdownFiles(p);
		else if (e.isFile() && e.name.endsWith(".md")) yield p;
	}
}

/**
 * @param {string} filePath
 * @param {string} scanRoot resolved
 */
export function relPath(filePath, scanRoot) {
	return relative(scanRoot, filePath);
}
