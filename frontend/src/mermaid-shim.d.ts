// Minimal ambient declaration for the `mermaid` npm package.
// mermaid 11 ships TypeScript 5+ types that pull in @types/d3-dispatch,
// whose `const` type parameters do not parse under the project's TypeScript 4.6.
// This shim covers only the API surface used by md-preview.
declare module "mermaid" {
	export interface MermaidConfig {
		[key: string]: unknown;
	}

	export interface RenderResult {
		svg: string;
		bindFunctions?: (element: Element) => void;
	}

	export interface RunOptions {
		querySelector?: string;
		nodes?: ArrayLike<HTMLElement>;
		postRenderCallback?: (id: string) => unknown;
		suppressErrors?: boolean;
	}

	export interface Mermaid {
		startOnLoad: boolean;
		initialize(config: MermaidConfig): void;
		render(id: string, code: string): Promise<RenderResult>;
		run(options?: RunOptions): Promise<void>;
	}

	const mermaid: Mermaid;
	export default mermaid;
}
