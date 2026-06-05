import { useEffect, useState } from "react";
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

const statusUnknown = "Loading preview...";

function App() {
	const [payload, setPayload] = useState<PreviewPayload>({
		filePath: "",
		html: "",
		version: "",
		renderedAt: "",
	});
	const [busy, setBusy] = useState(true);

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

	return (
		<div className="min-h-screen bg-slate-100 text-slate-900">
			<div className="mx-auto flex max-w-[1100px] flex-col gap-4 p-4">
				<header className="rounded-xl border border-slate-200/80 bg-white/95 p-4 shadow-sm shadow-slate-900/5 backdrop-blur">
					<h1 className="truncate text-lg font-semibold text-slate-900">
						{payload.filePath || "Markdown Preview"}
					</h1>
					<p className="mt-1 text-sm text-slate-500">
						{busy ? statusUnknown : `Preview version ${payload.version || "N/A"}`}
						{payload.renderedAt ? ` Updated ${payload.renderedAt}` : ""}
					</p>
				</header>

				{payload.error ? (
					<div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
						{payload.error}
					</div>
				) : null}

				<section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
					<div
						className="markdown-body"
						dangerouslySetInnerHTML={{ __html: payload.html || "<p>No preview content.</p>" }}
					/>
				</section>
			</div>
		</div>
	);
}

export default App;
