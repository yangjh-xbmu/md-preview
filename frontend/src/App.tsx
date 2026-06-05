import { ChangeEvent, MouseEvent, useEffect, useState } from "react";
import { EventsOff, EventsOn } from "../wailsjs/runtime";
import { LoadMarkdown } from "../wailsjs/go/main/App";
import "github-markdown-css/github-markdown.css";
import "./App.css";

type PreviewPayload = {
	filePath: string;
	html: string;
	version: string;
	renderedAt: string;
	error?: string;
};

type TocItem = {
	id: string;
	text: string;
	level: number;
};

type ThemeName = "github-light" | "github-dark" | "github-sepia";

type Theme = {
	name: ThemeName;
	label: string;
	description: string;
};

const fallbackMarkup = "<p>No preview content.</p>";

const themes: Theme[] = [
	{ name: "github-light", label: "GitHub Light", description: "Default style, close to the GitHub light reading theme." },
	{ name: "github-dark", label: "GitHub Dark", description: "Dark mode with stronger contrast for low-light conditions." },
	{ name: "github-sepia", label: "GitHub Sepia", description: "Warm, paper-like theme for long-form reading." },
];

const statusUnknown = "Loading preview...";
const themeStorageKey = "md-preview.theme";

function slugifyHeading(text: string, fallbackIndex: number): string {
	const base = text
		.trim()
		.toLowerCase()
		.replace(/[^\w\u4e00-\u9fa5]+/g, "-")
		.replace(/^-+|-+$/g, "")
		.replace(/-+/g, "-");

	if (!base) {
		return `heading-${fallbackIndex}`;
	}
	return base;
}

function extractTocAndNormalizeHtml(rawHtml: string): { html: string; toc: TocItem[] } {
	if (!rawHtml) {
		return { html: fallbackMarkup, toc: [] };
	}

	const parser = new DOMParser();
	const doc = parser.parseFromString(rawHtml, "text/html");
	const used = new Map<string, number>();
	const headings = Array.from(doc.querySelectorAll("h1, h2, h3, h4, h5, h6"));
	const toc: TocItem[] = [];

	headings.forEach((heading, index) => {
		const idSource = heading.getAttribute("id") || heading.textContent || "";
		let id = heading.getAttribute("id");
		if (!id || !id.trim()) {
			id = slugifyHeading(idSource, index);
		}

		const unique = `${id}-${(used.get(id) || 0) + 1}`;
		const finalId = used.get(id) ? unique : id;
		used.set(id, (used.get(id) || 0) + 1);
		heading.setAttribute("id", finalId);

		const text = heading.textContent?.trim() || "";
		const level = Number(heading.tagName.substring(1));

		if (text) {
			toc.push({
				id: finalId,
				text,
				level,
			});
		}
	});

	return {
		html: doc.body ? doc.body.innerHTML : rawHtml,
		toc,
	};
}

function App() {
	const [payload, setPayload] = useState<PreviewPayload>({
		filePath: "",
		html: "",
		version: "",
		renderedAt: "",
	});
	const [contentHtml, setContentHtml] = useState(fallbackMarkup);
	const [busy, setBusy] = useState(true);
	const [theme, setTheme] = useState<ThemeName>(() => {
		const saved = window.localStorage.getItem(themeStorageKey) as ThemeName | null;
		return saved === "github-dark" || saved === "github-sepia" || saved === "github-light" ? saved : "github-light";
	});
	const [toc, setToc] = useState<TocItem[]>([]);

	const applyPayload = (next: PreviewPayload) => {
		setPayload(next);
		if (next.error) {
			setContentHtml(fallbackMarkup);
			setToc([]);
			return;
		}

		const normalized = extractTocAndNormalizeHtml(next.html || "");
		setContentHtml(normalized.html);
		setToc(normalized.toc);
	};

	useEffect(() => {
		let mounted = true;

		const loadInitial = async () => {
			try {
				const next = await LoadMarkdown();
				if (!mounted) {
					return;
				}
				applyPayload(next);
			} catch {
				if (mounted) {
					setPayload((current) => ({
						...current,
						error: "Failed to load Markdown preview. Check the file path and permissions.",
					}));
					setContentHtml(fallbackMarkup);
					setToc([]);
				}
			} finally {
				if (mounted) {
					setBusy(false);
				}
			}
		};

		const handleUpdate = (next: PreviewPayload) => {
			if (mounted) {
				applyPayload(next);
			}
		};

		loadInitial();
		EventsOn("markdown-updated", handleUpdate);

		return () => {
			mounted = false;
			EventsOff("markdown-updated");
		};
	}, []);

	useEffect(() => {
		document.documentElement.setAttribute("data-theme", theme);
		window.localStorage.setItem(themeStorageKey, theme);
	}, [theme]);

	const onThemeChange = (event: ChangeEvent<HTMLSelectElement>) => {
		setTheme(event.target.value as ThemeName);
	};

	const onTocNavigation = (event: MouseEvent<HTMLAnchorElement>, id: string) => {
		event.preventDefault();
		const target = document.getElementById(id);
		if (target) {
			target.scrollIntoView({ behavior: "smooth", block: "start" });
		}
	};

	return (
		<div className={`min-h-screen transition-colors duration-200 md-preview-root theme-shell-${theme}`}>
			<div className="mx-auto flex max-w-[1400px] flex-col gap-4 p-4">
				<header className="md-preview-header">
					<h1 className="truncate text-lg font-semibold md-preview-title">
						{payload.filePath || "Markdown Preview"}
					</h1>
					<p className="mt-1 text-sm md-preview-subtle">
						{busy ? statusUnknown : `Preview version ${payload.version || "N/A"}`}
						{payload.renderedAt ? ` Updated ${payload.renderedAt}` : ""}
					</p>
					<div className="mt-3 flex items-center gap-2">
						<label className="text-sm md-preview-subtle" htmlFor="theme-select">
							Theme
						</label>
						<select
							id="theme-select"
							value={theme}
							onChange={onThemeChange}
							className="rounded-md border px-2 py-1 text-sm md-preview-select"
						>
							{themes.map((item) => (
								<option key={item.name} value={item.name}>
									{item.label}
								</option>
							))}
						</select>
						<p className="text-xs md-preview-subtle">
							{themes.find((item) => item.name === theme)?.description}
						</p>
					</div>
				</header>

				{payload.error ? (
					<div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
						{payload.error}
					</div>
				) : null}

				<div className="md-preview-content-layout">
					<section className="md-preview-panel md-preview-main-panel">
						<div className={`markdown-body theme-${theme}`} dangerouslySetInnerHTML={{ __html: contentHtml }} />
					</section>

					{toc.length > 0 ? (
						<aside className="md-preview-toc md-preview-panel">
							<h2 className="mb-3 text-sm font-semibold">Table of Contents</h2>
							<nav aria-label="Table of contents">
								<ul className="space-y-1">
									{toc.map((item) => (
										<li key={item.id} style={{ paddingLeft: `${Math.max(item.level - 1, 0) * 0.75}rem` }}>
											<a
												href={`#${item.id}`}
												onClick={(event) => onTocNavigation(event, item.id)}
												className="md-preview-subtle text-sm"
											>
												{item.text}
											</a>
										</li>
									))}
								</ul>
							</nav>
						</aside>
					) : null}
				</div>
			</div>
		</div>
	);
}

export default App;
