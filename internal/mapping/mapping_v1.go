package mapping

type KeyMode_V1 int

const (
	Keyboard_V1 KeyMode_V1 = iota
	Mouse_V1
)

type KeyMapping_V1 struct {
	Code int        `json:"code"`
	Mode KeyMode_V1 `json:"mode"`
}

type AxisMapping_V1 struct {
	PositiveKey KeyMapping_V1 `json:"positive_key"`
	NegativeKey KeyMapping_V1 `json:"negative_key"`
}

type HatMapping_V1 struct {
	Up    KeyMapping_V1 `json:"up"`
	Down  KeyMapping_V1 `json:"down"`
	Left  KeyMapping_V1 `json:"left"`
	Right KeyMapping_V1 `json:"right"`
}

type WindowProfileCfg_V1 struct {
	NamePattern  string `json:"name,omitempty"`
	ClassPattern string `json:"class,omitempty"`
}

type Mapping_V1 struct {
	WindowProfile WindowProfileCfg_V1      `json:"window_profiles"`
	AxisDeadzone  int16                    `json:"axes_deadzone,omitempty"`
	Axes          map[uint8]AxisMapping_V1 `json:"axes,omitempty"`
	Buttons       map[uint8]KeyMapping_V1  `json:"buttons,omitempty"`
	Hats          map[uint8]HatMapping_V1  `json:"hats,omitempty"`
}
