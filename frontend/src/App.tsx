import { ChangeEvent, KeyboardEvent, MouseEvent, useEffect, useRef, useState } from "react";
import { EventsOff, EventsOn, OnFileDrop, OnFileDropOff } from "../wailsjs/runtime";
import { ExportHTML, LoadMarkdown, OpenMarkdownFile, PrintPreview, SetFile } from "../wailsjs/go/main/App";
import "github-markdown-css/github-markdown.css";
import Prism from "prismjs";
import "prismjs/components/prism-markup";
import "prismjs/components/prism-clike";
import "prismjs/components/prism-c";
import "prismjs/components/prism-cpp";
import "prismjs/components/prism-css";
import "prismjs/components/prism-go";
import "prismjs/components/prism-json";
import "prismjs/components/prism-javascript";
import "prismjs/components/prism-python";
import "prismjs/components/prism-sql";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-yaml";
import "prismjs/components/prism-markdown";
import "prismjs/components/prism-diff";
import "prismjs/themes/prism.css";
import "prismjs/plugins/line-numbers/prism-line-numbers";
import "prismjs/plugins/line-numbers/prism-line-numbers.css";
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

type CodeBlockLanguage = string;

const fallbackMarkup = "<p>No Markdown file is loaded. Open a file or pass one on the command line.</p>";

const themes: Theme[] = [
	{ name: "github-light", label: "GitHub Light", description: "Default style, close to the GitHub light reading theme." },
	{ name: "github-dark", label: "GitHub Dark", description: "Dark mode with stronger contrast for low-light conditions." },
	{ name: "github-sepia", label: "GitHub Sepia", description: "Warm, paper-like theme for long-form reading." },
];

const statusUnknown = "Loading preview...";
const themeStorageKey = "md-preview.theme";
const backendLoadTimeoutMs = 3000;

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

function readCodeLanguage(codeBlock: HTMLElement): CodeBlockLanguage | null {
	const className = codeBlock.getAttribute("class") || "";
	const match = className.match(/language-([\w-]+)/i);
	if (!match) {
		return null;
	}
	return match[1].toLowerCase();
}

function getLanguageLabel(codeBlock: HTMLElement): string {
	const language = readCodeLanguage(codeBlock);
	return language ? language.toUpperCase() : "";
}

function defaultExportPath(filePath: string): string {
	const fallback = "document-preview.html";
	const trimmed = filePath.trim();
	if (!trimmed) {
		return fallback;
	}

	const base = trimmed.replace(/\.markdown$/i, "").replace(/\.md$/i, "");
	return `${base}-preview.html`;
}

function withTimeout<T>(promise: Promise<T>, timeoutMs: number, message: string): Promise<T> {
	return new Promise((resolve, reject) => {
		const timeout = window.setTimeout(() => reject(new Error(message)), timeoutMs);
		promise
			.then((value) => resolve(value))
			.catch((error) => reject(error))
			.finally(() => window.clearTimeout(timeout));
	});
}

function App() {
	const [payload, setPayload] = useState<PreviewPayload>({
		filePath: "",
		html: "",
		version: "",
		renderedAt: "",
	});
	const [filePathInput, setFilePathInput] = useState("");
	const [contentHtml, setContentHtml] = useState(fallbackMarkup);
	const [busy, setBusy] = useState(true);
	const [theme, setTheme] = useState<ThemeName>(() => {
		const saved = window.localStorage.getItem(themeStorageKey) as ThemeName | null;
		return saved === "github-dark" || saved === "github-sepia" || saved === "github-light" ? saved : "github-light";
	});
	const [toc, setToc] = useState<TocItem[]>([]);
	const [actionMessage, setActionMessage] = useState("");
	const previewRef = useRef<HTMLDivElement | null>(null);

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
				const next = await withTimeout(
					LoadMarkdown(),
					backendLoadTimeoutMs,
					"Backend did not respond. Use Open File or restart the app with a Markdown path.",
				);
				if (!mounted) {
					return;
				}
				applyPayload(next);
			} catch (error) {
				if (mounted) {
					const message = error instanceof Error ? error.message : "Failed to load Markdown preview.";
					setPayload((current) => ({
						...current,
						error: message,
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
		const onDrop = (_x: number, _y: number, paths: string[]) => {
			const first = paths && paths[0] ? paths[0].trim() : "";
			if (!first) {
				return;
			}
			setFilePathInput(first);
			void loadFromPath(first);
		};

		OnFileDrop(onDrop, false);

		return () => {
			OnFileDropOff();
		};
	}, []);

	useEffect(() => {
		document.documentElement.setAttribute("data-theme", theme);
		window.localStorage.setItem(themeStorageKey, theme);
	}, [theme]);

	useEffect(() => {
		const root = previewRef.current;
		if (!root) {
			return;
		}

		const blocks = Array.from(root.querySelectorAll("pre > code")).map((node) => node as HTMLElement);
		blocks.forEach((codeBlock) => {
			const pre = codeBlock.parentElement;
			if (!pre) {
				return;
			}

			pre.classList.add("line-numbers", "md-code-block");
			if (!pre.querySelector(".md-code-copy")) {
				const copyButton = document.createElement("button");
				const language = getLanguageLabel(codeBlock);
				copyButton.type = "button";
				copyButton.className = "md-code-copy";
				copyButton.textContent = language ? `${language} Copy` : "Copy";
				pre.appendChild(copyButton);

				copyButton.addEventListener("click", async () => {
					try {
						await navigator.clipboard.writeText(codeBlock.textContent || "");
						copyButton.textContent = "Copied";
						window.setTimeout(() => {
							copyButton.textContent = language ? `${language} Copy` : "Copy";
						}, 1200);
					} catch {
						copyButton.textContent = "Failed";
						window.setTimeout(() => {
							copyButton.textContent = language ? `${language} Copy` : "Copy";
						}, 1200);
					}
				});
			}
		});

		Prism.highlightAllUnder(root);
	}, [contentHtml]);

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

	const onFilePathChange = (event: ChangeEvent<HTMLInputElement>) => {
		setFilePathInput(event.target.value);
	};

	const loadFromPath = async (pathValue?: string) => {
		const trimmed = (pathValue ?? filePathInput).trim();
		if (!trimmed) {
			setActionMessage("Please enter a Markdown file path.");
			return;
		}

		try {
			const next = await SetFile(trimmed);
			applyPayload(next);
			setActionMessage(next.error ? next.error : "File loaded.");
		} catch {
			setActionMessage("Failed to load file path. Check the path and permissions.");
		}
	};

	const openPathOnEnter = (event: KeyboardEvent<HTMLInputElement>) => {
		if (event.key === "Enter") {
			loadFromPath();
		}
	};

	const openMarkdownFile = async () => {
		try {
			const next = await OpenMarkdownFile();
			applyPayload(next);
			setFilePathInput(next.filePath || filePathInput);
			setActionMessage(next.error ? next.error : "File loaded.");
		} catch {
			setActionMessage("Failed to open Markdown file.");
		}
	};

	const exportCurrentHtml = async () => {
		if (!payload.filePath) {
			setActionMessage("Open a Markdown file before exporting.");
			return;
		}

		const suggestedPath = defaultExportPath(payload.filePath);
		const target = window.prompt("Export HTML to path", suggestedPath);
		if (target === null || !target.trim()) {
			return;
		}

		try {
			const saved = await ExportHTML(target, theme);
			setActionMessage(`Exported HTML to: ${saved}`);
		} catch (err) {
			const message = err instanceof Error ? err.message : "Export failed";
			setActionMessage(message);
		}
	};

	const printToPDF = async () => {
		try {
			await PrintPreview();
			setActionMessage("Print dialog opened. Choose Save as PDF to export PDF.");
		} catch (err) {
			const message = err instanceof Error ? err.message : "Print failed";
			setActionMessage(message);
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
					<div className="mt-3 flex flex-wrap items-center gap-2">
						<button
							type="button"
							onClick={exportCurrentHtml}
							className="rounded-md border px-3 py-1.5 text-sm font-medium md-preview-select"
						>
							Export HTML
						</button>
						<button
							type="button"
							onClick={printToPDF}
							className="rounded-md border px-3 py-1.5 text-sm font-medium md-preview-select"
						>
							Export PDF
						</button>
					</div>
					<div className="mt-3 flex flex-wrap items-center gap-2">
						<button
							type="button"
							onClick={openMarkdownFile}
							className="rounded-md border px-3 py-1.5 text-sm font-medium md-preview-select"
						>
							Open File
						</button>
						<input
							value={filePathInput}
							onChange={onFilePathChange}
							onKeyDown={openPathOnEnter}
							placeholder="Paste Markdown file path and press Enter"
							type="text"
							className="min-w-[280px] flex-1 rounded-md border bg-transparent px-3 py-1.5 text-sm md-preview-select"
						/>
						<button
							type="button"
							onClick={() => loadFromPath()}
							className="rounded-md border px-3 py-1.5 text-sm font-medium md-preview-select"
						>
							Load File
						</button>
					</div>
					{actionMessage ? <p className="mt-2 text-xs md-preview-subtle">{actionMessage}</p> : null}
				</header>

				{payload.error ? (
					<div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
						{payload.error}
					</div>
				) : null}

				<div className="md-preview-content-layout">
					<section className="md-preview-panel md-preview-main-panel">
						<div ref={previewRef} className={`markdown-body theme-${theme}`} dangerouslySetInnerHTML={{ __html: contentHtml }} />
					</section>

					{toc.length > 0 ? (
						<aside className="md-preview-toc md-preview-panel">
							<h2 className="mb-3 text-sm font-semibold">Table of Contents</h2>
							<nav aria-label="Table of contents">
								<ul className="space-y-1">
									{toc.map((item) => (
										<li key={item.id} style={{ paddingLeft: `${Math.max(item.level - 1, 0) * 0.75}rem` }}>
											<a href={`#${item.id}`} onClick={(event) => onTocNavigation(event, item.id)} className="md-preview-subtle text-sm">
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
