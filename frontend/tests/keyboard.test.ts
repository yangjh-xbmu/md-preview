import { isFullscreenShortcut, isQuitShortcut } from "../src/keyboard";

type ShortcutEvent = {
	key: string;
	altKey: boolean;
	ctrlKey: boolean;
	metaKey: boolean;
	shiftKey: boolean;
};

const event = (overrides: Partial<ShortcutEvent> = {}): ShortcutEvent => ({
	key: "F11",
	altKey: false,
	ctrlKey: false,
	metaKey: false,
	shiftKey: false,
	...overrides,
});

const assert = (condition: boolean, message: string) => {
	if (!condition) {
		throw new Error(message);
	}
};

assert(isFullscreenShortcut(event()), "F11 should toggle full screen");
assert(isFullscreenShortcut(event({ altKey: true })), "Option+F11 should toggle full screen");
assert(!isFullscreenShortcut(event({ ctrlKey: true })), "Ctrl+F11 should not toggle full screen");
assert(!isFullscreenShortcut(event({ metaKey: true })), "Cmd+F11 should not toggle full screen");
assert(!isFullscreenShortcut(event({ shiftKey: true })), "Shift+F11 should not toggle full screen");

assert(isQuitShortcut(event({ key: "q", ctrlKey: true })), "Ctrl+Q should quit");
assert(isQuitShortcut(event({ key: "Q", ctrlKey: true })), "Ctrl+Q should normalize the key case");
assert(!isQuitShortcut(event({ key: "q", metaKey: true })), "Cmd+Q should not use the Ctrl+Q shortcut");
assert(!isQuitShortcut(event({ key: "q", ctrlKey: true, altKey: true })), "Ctrl+Alt+Q should not quit");
assert(!isQuitShortcut(event({ key: "q", ctrlKey: true, shiftKey: true })), "Ctrl+Shift+Q should not quit");

console.log("keyboard shortcuts: pass");
