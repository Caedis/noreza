package mapping

// Maps KeyboardEvent.key strings to Linux input-event codes
var KeyToCode = map[string]int{
	// unset
	"": 0,
	// Letters
	"a": 30, "b": 48, "c": 46, "d": 32, "e": 18, "f": 33,
	"g": 34, "h": 35, "i": 23, "j": 36, "k": 37, "l": 38,
	"m": 50, "n": 49, "o": 24, "p": 25, "q": 16, "r": 19,
	"s": 31, "t": 20, "u": 22, "v": 47, "w": 17, "x": 45,
	"y": 21, "z": 44,

	// Numbers
	"1": 2, "2": 3, "3": 4, "4": 5, "5": 6,
	"6": 7, "7": 8, "8": 9, "9": 10, "0": 11,

	// Function keys
	"F1": 59, "F2": 60, "F3": 61, "F4": 62, "F5": 63,
	"F6": 64, "F7": 65, "F8": 66, "F9": 67, "F10": 68,
	"F11": 87, "F12": 88,

	// Modifiers & controls
	"Escape": 1, "Tab": 15, "CapsLock": 58, "Shift": 42, "Control": 29, "Alt": 56,
	" ": 57, "Enter": 28, "Backspace": 14,

	// Arrows
	"ArrowUp": 103, "ArrowDown": 108, "ArrowLeft": 105, "ArrowRight": 106,

	// Navigation
	"Insert": 110, "Delete": 111, "Home": 102, "End": 107, "PageUp": 104, "PageDown": 109,

	// Symbols (common)
	"-": 12, "=": 13, "[": 26, "]": 27,
	";": 39, "'": 40, "`": 41, "\\": 43,
	",": 51, ".": 52, "/": 53,

	// Numpad keys
	"NumLock": 69, "NumpadDivide": 98, "NumpadMultiply": 55, "NumpadSubtract": 74,
	"NumpadAdd": 78, "NumpadEnter": 96, "Numpad1": 79, "Numpad2": 80, "Numpad3": 81,
	"Numpad4": 75, "Numpad5": 76, "Numpad6": 77, "Numpad7": 71, "Numpad8": 72, "Numpad9": 73,
	"Numpad0": 82, "NumpadDecimal": 83,
}

var MouseToCode = map[string]int{
	// Mouse buttons
	"LClick": 0x110, "RClick": 0x111, "MClick": 0x112,
}

// Maps Linux input-event codes to KeyboardEvent.key
var CodeToKey = func() map[int]string {
	m := make(map[int]string, len(KeyToCode))
	for k, v := range KeyToCode {
		if k == " " {
			m[v] = "Space"
		} else {
			m[v] = k
		}
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
