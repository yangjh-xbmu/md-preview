import { useEffect, useRef, useState } from "react";
import {
	ClipboardSetText,
	EventsOff,
	EventsOn,
	OnFileDrop,
	OnFileDropOff,
	WindowFullscreen,
	WindowIsFullscreen,
	WindowUnfullscreen,
} from "../wailsjs/runtime";
import {
	CheckForUpdates,
	ExportHTMLWithDialog,
	GetUpdateSettings,
	GetUpdateStatus,
	InstallStagedUpdate,
	LoadMarkdown,
	OpenMarkdownFile,
	PrintPreview,
	ResolveWikiLink,
	SetAutoUpdateEnabled,
	SetFile,
	SetTheme,
} from "../wailsjs/go/main/App";
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
import FrontmatterTable from "./FrontmatterTable";
import { renderMermaidBlocks, reinitForTheme } from "./mermaid";

type PreviewPayload = {
	filePath: string;
	html: string;
	version: string;
	renderedAt: string;
	error?: string;
	frontmatter?: Record<string, unknown> | null;
};

type TocItem = {
	id: string;
	text: string;
	level: number;
};

type ThemeName = "github-light" | "github-dark" | "github-sepia";

type CodeBlockLanguage = string;

type UpdateStatus = {
	state: string;
	currentVersion: string;
	latestVersion: string;
	message: string;
	downloadedPath: string;
	releaseURL: string;
	checkedAt: string;
};

const fallbackMarkup = "<p>No Markdown file is loaded. Open a file or pass one on the command line.</p>";

const statusUnknown = "Loading preview...";
const themeStorageKey = "md-preview.theme";
const backendLoadTimeoutMs = 3000;

const themeLabels: Record<ThemeName, string> = {
	"github-light": "Light",
	"github-dark": "Dark",
	"github-sepia": "Sepia",
};

const initialUpdateStatus: UpdateStatus = {
	state: "idle",
	currentVersion: "",
	latestVersion: "",
	message: "Automatic updates are enabled.",
	downloadedPath: "",
	releaseURL: "",
	checkedAt: "",
};

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
	const [contentHtml, setContentHtml] = useState(fallbackMarkup);
	const [frontmatterData, setFrontmatterData] = useState<Record<string, unknown> | null>(null);
	const [busy, setBusy] = useState(true);
	const [theme, setTheme] = useState<ThemeName>(() => {
		const saved = window.localStorage.getItem(themeStorageKey) as ThemeName | null;
		return saved === "github-dark" || saved === "github-sepia" || saved === "github-light" ? saved : "github-light";
	});
	const [toc, setToc] = useState<TocItem[]>([]);
	const [tocVisible, setTocVisible] = useState(true);
	const [fullscreen, setFullscreen] = useState(false);
	const [actionMessage, setActionMessage] = useState("");
	const [autoUpdateEnabled, setAutoUpdateEnabledState] = useState(true);
	const [updateStatus, setUpdateStatus] = useState<UpdateStatus>(initialUpdateStatus);
	const [menuOpen, setMenuOpen] = useState(false);
	const navHistory = useRef<string[]>([]);
	const navIndex = useRef(-1);
	const previewRef = useRef<HTMLDivElement | null>(null);
	const menuRef = useRef<HTMLDivElement | null>(null);

	const applyPayload = (next: PreviewPayload, addToHistory = true) => {
		setPayload(next);
		if (next.error) {
			setContentHtml(fallbackMarkup);
			setToc([]);
			setFrontmatterData(null);
			return;
		}

		const normalized = extractTocAndNormalizeHtml(next.html || "");
		setContentHtml(normalized.html);
		setToc(normalized.toc);
		setFrontmatterData(typeof next.frontmatter === "object" && next.frontmatter !== null ? next.frontmatter as Record<string, unknown> : null);

		if (addToHistory && next.filePath && !next.error) {
			const hist = navHistory.current;
			const idx = navIndex.current;
			if (idx >= 0 && idx < hist.length && hist[idx] === next.filePath) {
				return;
			}
			hist.splice(idx + 1);
			hist.push(next.filePath);
			navIndex.current = hist.length - 1;
		}
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
		const handleThemeChange = (nextTheme: ThemeName) => {
			if (mounted) {
				setTheme(nextTheme);
			}
		};
		const handleStatusMessage = (message: string) => {
			if (mounted) {
				setActionMessage(message);
			}
		};
		const handleUpdateStatus = (nextStatus: UpdateStatus) => {
			if (mounted) {
				setUpdateStatus(nextStatus);
			}
		};
		const syncFullscreenState = async () => {
			try {
				const isFullscreen = await WindowIsFullscreen();
				if (mounted) {
					setFullscreen(isFullscreen);
				}
			} catch {
				if (mounted) {
					setFullscreen(false);
				}
			}
		};

		loadInitial();
		void (async () => {
			try {
				const [settings, status] = await Promise.all([GetUpdateSettings(), GetUpdateStatus()]);
				if (mounted) {
					setAutoUpdateEnabledState(settings.autoUpdateEnabled);
					setUpdateStatus(status);
				}
			} catch {
				if (mounted) {
					setUpdateStatus({
						...initialUpdateStatus,
						state: "failed",
						message: "Failed to load update settings.",
					});
				}
			}
		})();
		void syncFullscreenState();
		EventsOn("markdown-updated", handleUpdate);
		EventsOn("theme-changed", handleThemeChange);
		EventsOn("status-message", handleStatusMessage);
		EventsOn("update-status-changed", handleUpdateStatus);

		return () => {
			mounted = false;
			EventsOff("markdown-updated");
			EventsOff("theme-changed");
			EventsOff("status-message");
			EventsOff("update-status-changed");
		};
	}, []);

	useEffect(() => {
		const onDrop = (_x: number, _y: number, paths: string[]) => {
			const first = paths && paths[0] ? paths[0].trim() : "";
			if (!first) {
				return;
			}
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
		void SetTheme(theme);
	}, [theme]);

	useEffect(() => {
		const onDocumentPointerDown = (event: PointerEvent) => {
			if (!menuRef.current || menuRef.current.contains(event.target as Node)) {
				return;
			}
			setMenuOpen(false);
		};
		const onKeyDown = (event: KeyboardEvent) => {
			if (event.key === "Escape") {
				setMenuOpen(false);
				return;
			}

			if (event.key === "F11") {
				event.preventDefault();
				void toggleFullscreen();
				return;
			}
			if (event.altKey && event.key === "ArrowLeft") {
				event.preventDefault();
				void navigateBack();
				return;
			}
			if (event.altKey && event.key === "ArrowRight") {
				event.preventDefault();
				void navigateForward();
				return;
			}

			if (!event.ctrlKey && !event.metaKey) {
				return;
			}

			const key = event.key.toLowerCase();
			if (key === "c") {
				const text = window.getSelection()?.toString() ?? "";
				if (text.trim()) {
					event.preventDefault();
					void ClipboardSetText(text);
				}
				return;
			}
			if (key === "a") {
				event.preventDefault();
				const selection = document.getSelection();
				if (selection && previewRef.current) {
					selection.selectAllChildren(previewRef.current);
				}
				return;
			}
			if (key === "o") {
				event.preventDefault();
				void openMarkdownFile();
			}
			if (key === "s") {
				event.preventDefault();
				void exportHtml();
			}
			if (key === "p") {
				event.preventDefault();
				printToPdf();
			}
			if (key === "t") {
				event.preventDefault();
				toggleToc();
			}
		};

		document.addEventListener("pointerdown", onDocumentPointerDown);
		window.addEventListener("keydown", onKeyDown);

		return () => {
			document.removeEventListener("pointerdown", onDocumentPointerDown);
			window.removeEventListener("keydown", onKeyDown);
		};
	});

	useEffect(() => {
		const root = previewRef.current;
		if (!root) {
			return;
		}

		reinitForTheme(theme);
		void renderMermaidBlocks(root, theme);
	}, [contentHtml, theme]);

	useEffect(() => {
		const root = previewRef.current;
		if (!root) {
			return;
		}

		const blocks = Array.from(root.querySelectorAll("pre > code")).map((node) => node as HTMLElement);
		blocks.forEach((codeBlock) => {
			if (codeBlock.classList.contains("language-mermaid")) {
				return;
			}
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

	useEffect(() => {
		const previewEl = previewRef.current;
		if (!previewEl) return;

		const handleMouseUp = (event: MouseEvent) => {
			if (event.button !== 0) return;
			if ((event.target as HTMLElement).closest(".md-code-copy")) return;

			const text = window.getSelection()?.toString() ?? "";
			if (!text.trim()) return;

			void ClipboardSetText(text);
		};

		previewEl.addEventListener("mouseup", handleMouseUp);
		return () => previewEl.removeEventListener("mouseup", handleMouseUp);
	}, []);

	useEffect(() => {
			const previewEl = previewRef.current;
			if (!previewEl) return;

			const handleClick = (event: MouseEvent) => {
				const anchor = (event.target as HTMLElement).closest("a") as HTMLAnchorElement | null;
				if (!anchor) return;

				const href = anchor.getAttribute("href") || "";
				if (!href || href.startsWith("#") || href.startsWith("http://") || href.startsWith("https://") || href.startsWith("mailto:")) return;

				event.preventDefault();

				void (async () => {
					try {
						const resolved = await ResolveWikiLink(href);
						if (!resolved) {
							setActionMessage(`Wiki link target not found: ${href}`);
							return;
						}
						const next = await SetFile(resolved);
						applyPayload(next);
						setActionMessage(next.error ? next.error : `Navigated to ${resolved}  (Alt+← back, Alt+→ forward)`);
					} catch {
						setActionMessage(`Failed to resolve wiki link: ${href}`);
					}
				})();
			};

			previewEl.addEventListener("click", handleClick);
			return () => previewEl.removeEventListener("click", handleClick);
		}, [contentHtml]);

	const onTocNavigation = (event: MouseEvent, id: string) => {
		event.preventDefault();
		const target = document.getElementById(id);
		if (target) {
			target.scrollIntoView({ behavior: "smooth", block: "start" });
		}
	};


		const navigateBack = async () => {
			const hist = navHistory.current;
			const idx = navIndex.current;
			if (idx <= 0) {
				setActionMessage("No previous page.");
				return;
			}
			navIndex.current = idx - 1;
			try {
				const next = await SetFile(hist[navIndex.current]);
				applyPayload(next, false);
				setActionMessage("Navigated back  (Alt+← back, Alt+→ forward)");
			} catch {
				setActionMessage("Failed to navigate back.");
			}
		};

		const navigateForward = async () => {
			const hist = navHistory.current;
			const idx = navIndex.current;
			if (idx >= hist.length - 1) {
				setActionMessage("No next page.");
				return;
			}
			navIndex.current = idx + 1;
			try {
				const next = await SetFile(hist[navIndex.current]);
				applyPayload(next, false);
				setActionMessage("Navigated forward  (Alt+← back, Alt+→ forward)");
			} catch {
				setActionMessage("Failed to navigate forward.");
			}
		};
	const loadFromPath = async (pathValue: string) => {
		const trimmed = pathValue.trim();
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

	const openMarkdownFile = async () => {
		setMenuOpen(false);
		try {
			const next = await OpenMarkdownFile();
			applyPayload(next);
			setActionMessage(next.error ? next.error : "File loaded.");
		} catch {
			setActionMessage("Failed to open Markdown file.");
		}
	};

	const exportHtml = async () => {
		setMenuOpen(false);
		try {
			const saved = await ExportHTMLWithDialog();
			if (saved) {
				setActionMessage(`Exported HTML to: ${saved}`);
			}
		} catch (error) {
			const message = error instanceof Error ? error.message : "Export failed.";
			setActionMessage(message);
		}
	};

	const printToPdf = () => {
		setMenuOpen(false);
		document.documentElement.classList.add("printing");
		PrintPreview();
		document.documentElement.classList.remove("printing");
		setActionMessage("Print dialog opened.");
	};

	const toggleToc = () => {
		setMenuOpen(false);
		setTocVisible((current) => {
			const next = !current;
			setActionMessage(next ? "Table of Contents shown." : "Table of Contents hidden.");
			return next;
		});
	};

	const toggleFullscreen = async () => {
		setMenuOpen(false);
		try {
			const isFullscreen = await WindowIsFullscreen();
			if (isFullscreen) {
				WindowUnfullscreen();
				setFullscreen(false);
				setActionMessage("Exited full screen.");
				return;
			}

			WindowFullscreen();
			setFullscreen(true);
			setActionMessage("Entered full screen.");
		} catch {
			setActionMessage("Failed to toggle full screen.");
		}
	};

	const selectTheme = (nextTheme: ThemeName) => {
		setMenuOpen(false);
		setTheme(nextTheme);
	};

	const toggleAutoUpdate = async () => {
		try {
			const next = await SetAutoUpdateEnabled(!autoUpdateEnabled);
			setAutoUpdateEnabledState(next.autoUpdateEnabled);
			const status = await GetUpdateStatus();
			setUpdateStatus(status);
			setActionMessage(status.message);
		} catch {
			setActionMessage("Failed to update automatic update setting.");
		}
	};

	const checkForUpdates = async () => {
		try {
			const status = await CheckForUpdates(true);
			setUpdateStatus(status);
			setActionMessage(status.message);
		} catch {
			setActionMessage("Failed to check for updates.");
		}
	};

	const installStagedUpdate = async () => {
		try {
			const status = await InstallStagedUpdate();
			setUpdateStatus(status);
			setActionMessage(status.message);
		} catch {
			setActionMessage("Failed to install staged update.");
		}
	};

	const showToc = toc.length > 0 && tocVisible;

	return (
		<div className={`min-h-screen transition-colors duration-200 md-preview-root theme-shell-${theme}`}>
			<div ref={menuRef} className="md-floating-menu">
				<button
					type="button"
					className="md-menu-trigger"
					aria-haspopup="menu"
					aria-expanded={menuOpen}
					onClick={() => setMenuOpen((current) => !current)}
				>
					Menu
				</button>

				{menuOpen ? (
					<div className="md-menu-popover" role="menu">
						<div className="md-menu-section">
							<p className="md-menu-label">File</p>
							<button type="button" role="menuitem" className="md-menu-item" onClick={openMarkdownFile}>
								<span>Open Markdown</span>
								<kbd>Ctrl O</kbd>
							</button>
							<button type="button" role="menuitem" className="md-menu-item" onClick={exportHtml}>
								<span>Export HTML</span>
								<kbd>Ctrl S</kbd>
							</button>
							<button type="button" role="menuitem" className="md-menu-item" onClick={printToPdf}>
								<span>Print / PDF</span>
								<kbd>Ctrl P</kbd>
							</button>
						</div>

						<div className="md-menu-section">
							<p className="md-menu-label">View</p>
							<button type="button" role="menuitem" className="md-menu-item" onClick={toggleToc}>
								<span>{tocVisible ? "Hide TOC" : "Show TOC"}</span>
								<kbd>Ctrl T</kbd>
							</button>
							<button type="button" role="menuitem" className="md-menu-item" onClick={toggleFullscreen}>
								<span>{fullscreen ? "Exit Full Screen" : "Full Screen"}</span>
								<kbd>F11</kbd>
							</button>
								<button type="button" role="menuitem" className="md-menu-item" onClick={() => { setMenuOpen(false); void navigateBack(); }}>
									<span>Back</span>
									<kbd>Alt ←</kbd>
								</button>
								<button type="button" role="menuitem" className="md-menu-item" onClick={() => { setMenuOpen(false); void navigateForward(); }}>
									<span>Forward</span>
									<kbd>Alt →</kbd>
								</button>
						</div>

						<div className="md-menu-section">
							<p className="md-menu-label">Theme</p>
							<div className="md-theme-grid">
								{(["github-light", "github-dark", "github-sepia"] as ThemeName[]).map((item) => (
									<button
										key={item}
										type="button"
										className={`md-theme-choice ${theme === item ? "is-active" : ""}`}
										onClick={() => selectTheme(item)}
									>
										<span className={`md-theme-swatch ${item}`} />
										<span>{themeLabels[item]}</span>
									</button>
								))}
							</div>
						</div>

						<div className="md-menu-section">
							<p className="md-menu-label">Updates</p>
							<button
								type="button"
								role="menuitemcheckbox"
								aria-checked={autoUpdateEnabled}
								className="md-menu-item"
								onClick={toggleAutoUpdate}
							>
								<span>Auto Updates</span>
								<span className={`md-update-pill ${autoUpdateEnabled ? "is-on" : "is-off"}`}>
									{autoUpdateEnabled ? "On" : "Off"}
								</span>
							</button>
							<button type="button" role="menuitem" className="md-menu-item" onClick={checkForUpdates}>
								<span>Check Updates</span>
								<span className="md-update-pill">{updateStatus.state === "checking" ? "..." : "Now"}</span>
							</button>
							{updateStatus.state === "ready" ? (
								<button type="button" role="menuitem" className="md-menu-item" onClick={installStagedUpdate}>
									<span>Restart to Install</span>
									<span className="md-update-pill is-on">Ready</span>
								</button>
							) : null}
							<div className={`md-update-status is-${updateStatus.state}`}>
								{updateStatus.message || "Automatic updates are enabled."}
							</div>
						</div>
					</div>
				) : null}
			</div>

			<div className="mx-auto flex max-w-[1400px] flex-col gap-3 p-4">
				{busy || actionMessage ? (
					<div className="md-preview-status md-preview-subtle">
						{busy ? statusUnknown : actionMessage}
					</div>
				) : null}

				{payload.error ? (
					<div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
						{payload.error}
					</div>
				) : null}

				<div className={`md-preview-content-layout ${showToc ? "" : "is-single-column"}`}>
					<section className="md-preview-panel md-preview-main-panel">
						<FrontmatterTable frontmatter={frontmatterData} />
						<div ref={previewRef} className={`markdown-body theme-${theme}`} dangerouslySetInnerHTML={{ __html: contentHtml }} />
					</section>

					{showToc ? (
						<aside className="md-preview-toc md-preview-panel">
							<h2 className="mb-3 text-sm font-semibold">Table of Contents</h2>
							<nav aria-label="Table of contents">
								<ul className="space-y-1">
									{toc.map((item) => (
										<li key={item.id} style={{ paddingLeft: `${Math.max(item.level - 1, 0) * 0.75}rem` }}>
											<a href={`#${item.id}`} onClick={(event) => onTocNavigation(event.nativeEvent, item.id)} className="md-preview-subtle text-sm">
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
