package mapping

// Maps KeyboardEvent.code strings to Linux input-event codes
var KeyToCode = map[string]int{
	// Letters
	"KeyA": 30, "KeyB": 48, "KeyC": 46, "KeyD": 32, "KeyE": 18, "KeyF": 33,
	"KeyG": 34, "KeyH": 35, "KeyI": 23, "KeyJ": 36, "KeyK": 37, "KeyL": 38,
	"KeyM": 50, "KeyN": 49, "KeyO": 24, "KeyP": 25, "KeyQ": 16, "KeyR": 19,
	"KeyS": 31, "KeyT": 20, "KeyU": 22, "KeyV": 47, "KeyW": 17, "KeyX": 45,
	"KeyY": 21, "KeyZ": 44,

	// Numbers
	"Digit1": 2, "Digit2": 3, "Digit3": 4, "Digit4": 5, "Digit5": 6,
	"Digit6": 7, "Digit7": 8, "Digit8": 9, "Digit9": 10, "Digit0": 11,

	// Function keys
	"F1": 59, "F2": 60, "F3": 61, "F4": 62, "F5": 63,
	"F6": 64, "F7": 65, "F8": 66, "F9": 67, "F10": 68,
	"F11": 87, "F12": 88,

	// Modifiers & controls
	"Escape": 1, "Tab": 15, "CapsLock": 58, "ShiftLeft": 42, "ShiftRight": 54,
	"ControlLeft": 29, "ControlRight": 97, "AltLeft": 56, "AltRight": 100,
	"Space": 57, "Enter": 28, "Backspace": 14,

	// Arrows
	"ArrowUp": 103, "ArrowDown": 108, "ArrowLeft": 105, "ArrowRight": 106,

	// Navigation
	"Insert": 110, "Delete": 111, "Home": 102, "End": 107,
	"PageUp": 104, "PageDown": 109,

	// Symbols
	"Minus": 12, "Equal": 13, "BracketLeft": 26, "BracketRight": 27,
	"Semicolon": 39, "Quote": 40, "Backquote": 41, "Backslash": 43,
	"Comma": 51, "Period": 52, "Slash": 53,

	// Numpad
	"NumLock": 69, "NumpadDivide": 98, "NumpadMultiply": 55, "NumpadSubtract": 74,
	"NumpadAdd": 78, "NumpadEnter": 96, "Numpad1": 79, "Numpad2": 80,
	"Numpad3": 81, "Numpad4": 75, "Numpad5": 76, "Numpad6": 77,
	"Numpad7": 71, "Numpad8": 72, "Numpad9": 73, "Numpad0": 82,
	"NumpadDecimal": 83,
}

var MouseToCode = map[string]int{
	"LClick": 0x110, "RClick": 0x111, "MClick": 0x112,
}

// Reverse map: Linux input-event code -> KeyboardEvent.code
var CodeToKey = func() map[int]string {
	m := make(map[int]string, len(KeyToCode))
	for k, v := range KeyToCode {
		m[v] = k
	}
	return m
}()

var CodeToMouse = func() map[int]string {
	m := make(map[int]string, len(MouseToCode))
	for k, v := range MouseToCode {
		m[v] = k
	}
	return m
}()
