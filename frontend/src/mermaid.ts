// INPUT: Markdown preview root element, active preview theme, Mermaid library.
// OUTPUT: Rendered Mermaid SVG diagrams placed into the preview DOM, or in-page error placeholders.
// POS: Frontend Mermaid rendering helper scoped to md-preview's preview surface.
import mermaid from "mermaid";

export type PreviewTheme = "github-light" | "github-dark" | "github-sepia";

export function mermaidThemeFor(theme: PreviewTheme): "default" | "dark" {
	return theme === "github-dark" ? "dark" : "default";
}

let initializedTheme: PreviewTheme | null = null;

function ensureInit(theme: PreviewTheme): void {
	if (initializedTheme === theme) {
		return;
	}
	mermaid.initialize({
		startOnLoad: false,
		theme: mermaidThemeFor(theme),
		securityLevel: "strict",
		fontFamily: "ui-monospace, SFMono-Regular, SF Mono, Menlo, Monaco, Consolas, Liberation Mono, Courier New, monospace",
	});
	initializedTheme = theme;
}

export function reinitForTheme(theme: PreviewTheme): void {
	ensureInit(theme);
}

function errorMessage(err: unknown): string {
	if (err instanceof Error) {
		return err.message;
	}
	if (err && typeof err === "object" && "str" in err) {
		const str = (err as { str: unknown }).str;
		return typeof str === "string" ? str : String(err);
	}
	return String(err);
}

/**
 * Render every `pre > code.language-mermaid` block in `root` as an inline SVG.
 * Re-renders existing `.md-mermaid` placeholders when called again with the same DOM
 * (e.g. on theme switch) by reading the preserved source from `data-mermaid-source`.
 *
 * Per-block errors are caught and surfaced in-place so sibling blocks still render.
 */
export async function renderMermaidBlocks(root: HTMLElement, theme: PreviewTheme): Promise<void> {
	ensureInit(theme);

	const codeBlocks = Array.from(root.querySelectorAll("pre > code.language-mermaid")) as HTMLElement[];
	const newPlaceholders: HTMLElement[] = [];

	for (const codeBlock of codeBlocks) {
		const pre = codeBlock.parentElement;
		if (!pre) {
			continue;
		}

		const source = codeBlock.textContent ?? "";
		const container = document.createElement("div");
		container.className = "md-mermaid";
		container.setAttribute("data-theme", theme);
		container.setAttribute("data-mermaid-source", source);
		container.setAttribute("role", "img");
		container.textContent = source;

		pre.replaceWith(container);
		newPlaceholders.push(container);
	}

	const existing = Array.from(root.querySelectorAll(".md-mermaid[data-mermaid-source]")) as HTMLElement[];
	const toRender = newPlaceholders.length > 0 ? newPlaceholders : existing;
	if (toRender.length === 0) {
		return;
	}

	await Promise.all(
		toRender.map(async (container, index) => {
			const source = container.getAttribute("data-mermaid-source") ?? "";

			container.className = "md-mermaid";
			container.setAttribute("data-theme", theme);
			container.setAttribute("role", "img");
			container.removeAttribute("aria-label");

			if (!source.trim()) {
				container.classList.add("md-mermaid-empty");
				container.textContent = "(empty Mermaid block)";
				return;
			}

			const id = `md-mermaid-svg-${index}-${Date.now()}`;
			try {
				const { svg } = await mermaid.render(id, source);
				container.innerHTML = svg;
			} catch (err) {
				container.classList.add("md-mermaid-error");
				container.setAttribute("role", "alert");
				container.textContent = `Mermaid render failed: ${errorMessage(err)}`;
			}
		}),
	);
}
