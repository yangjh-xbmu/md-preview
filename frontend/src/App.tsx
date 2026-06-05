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
	{ name: "github-light", label: "GitHub Light", description: "默认，接近 GitHub 官方阅读风格" },
	{ name: "github-dark", label: "GitHub Dark", description: "偏暗，适合低亮环境阅读" },
	{ name: "github-sepia", label: "Sepia", description: "纸张风格，适合长文阅读" },
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
		return saved === "github-dark" || saved === "github-sepia" ? saved : "github-light";
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
		<div className="min-h-screen bg-slate-100 text-slate-900 transition-colors duration-200">
			<div className="mx-auto flex max-w-[1100px] flex-col gap-4 p-4">
				<header className="rounded-xl border border-slate-200/80 bg-white/95 p-4 shadow-sm shadow-slate-900/5 backdrop-blur">
					<h1 className="truncate text-lg font-semibold text-slate-900">
						{payload.filePath || "Markdown Preview"}
					</h1>
					<p className="mt-1 text-sm text-slate-500">
						{busy ? statusUnknown : `Preview version ${payload.version || "N/A"}`}
						{payload.renderedAt ? ` Updated ${payload.renderedAt}` : ""}
					</p>
					<div className="mt-3 flex items-center gap-2">
						<label className="text-sm text-slate-600" htmlFor="theme-select">
							Theme
						</label>
						<select
							id="theme-select"
							value={theme}
							onChange={onThemeChange}
							className="rounded-md border border-slate-300 bg-white px-2 py-1 text-sm text-slate-900"
						>
							{themes.map((item) => (
								<option key={item.name} value={item.name}>
									{item.label}
								</option>
							))}
						</select>
						<p className="text-xs text-slate-500">{themes.find((item) => item.name === theme)?.description}</p>
					</div>
				</header>

				{payload.error ? (
					<div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
						{payload.error}
					</div>
				) : null}

				<section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
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
