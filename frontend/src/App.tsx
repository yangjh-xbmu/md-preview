import { ChangeEvent, useEffect, useState } from "react";
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

type ThemeName = "github-light" | "github-dark" | "github-sepia";

type Theme = {
	name: ThemeName;
	label: string;
	description: string;
};

const themes: Theme[] = [
	{ name: "github-light", label: "GitHub Light", description: "Default style, close to the GitHub light reading theme." },
	{ name: "github-dark", label: "GitHub Dark", description: "Dark mode with stronger contrast for low-light conditions." },
	{ name: "github-sepia", label: "GitHub Sepia", description: "Warm, paper-like theme for long-form reading." },
];

const statusUnknown = "Loading preview...";
const themeStorageKey = "md-preview.theme";

function App() {
	const [payload, setPayload] = useState<PreviewPayload>({
		filePath: "",
		html: "",
		version: "",
		renderedAt: "",
	});
	const [busy, setBusy] = useState(true);
	const [theme, setTheme] = useState<ThemeName>(() => {
		const saved = window.localStorage.getItem(themeStorageKey) as ThemeName | null;
		return saved === "github-dark" || saved === "github-sepia" || saved === "github-light" ? saved : "github-light";
	});

	useEffect(() => {
		let mounted = true;

		const loadInitial = async () => {
			try {
				const next = await LoadMarkdown();
				if (!mounted) {
					return;
				}
				setPayload(next);
			} catch {
				setPayload((current) => ({
					...current,
					error: "Failed to load Markdown preview. Check the file path and permissions.",
				}));
			} finally {
				if (mounted) {
					setBusy(false);
				}
			}
		};

		const handleUpdate = (next: PreviewPayload) => {
			if (mounted) {
				setPayload(next);
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

	return (
		<div className={`min-h-screen transition-colors duration-200 md-preview-root theme-shell-${theme}`}>
			<div className="mx-auto flex max-w-[1100px] flex-col gap-4 p-4">
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

				<section className="md-preview-panel">
					<div
						className={`markdown-body theme-${theme}`}
						dangerouslySetInnerHTML={{ __html: payload.html || "<p>No preview content.</p>" }}
					/>
				</section>
			</div>
		</div>
	);
}

export default App;
