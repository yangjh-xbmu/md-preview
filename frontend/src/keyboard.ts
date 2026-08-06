export type ShortcutEvent = Pick<KeyboardEvent, "key" | "altKey" | "ctrlKey" | "metaKey" | "shiftKey">;

export function isFullscreenShortcut(event: ShortcutEvent): boolean {
	return event.key === "F11" && !event.ctrlKey && !event.metaKey && !event.shiftKey;
}

export function isQuitShortcut(event: ShortcutEvent): boolean {
	return event.key.toLowerCase() === "q" && event.ctrlKey && !event.altKey && !event.metaKey && !event.shiftKey;
}
