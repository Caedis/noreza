package mapping

import (
	"encoding/json"
	"os"
)

type JoystickEvent struct {
	Type  string `json:"type"`
	Index uint8  `json:"index"`
	Value int16  `json:"value"`
}

func (j *JoystickEvent) String() string {
	str, _ := json.Marshal(j)
	return string(str)
}

type KeyMode int

const (
	Keyboard KeyMode = iota
	Mouse
)

type KeyMapping struct {
	Code int     `json:"code"`
	Mode KeyMode `json:"mode"`
}

type AxisMapping struct {
	PositiveKey KeyMapping `json:"positive_key"`
	NegativeKey KeyMapping `json:"negative_key"`
}

type HatMapping struct {
	Up    KeyMapping `json:"up"`
	Down  KeyMapping `json:"down"`
	Left  KeyMapping `json:"left"`
	Right KeyMapping `json:"right"`
}

type WindowProfileCfg struct {
	NamePattern  string `json:"name,omitempty"`
	ClassPattern string `json:"class,omitempty"`
}

type Mapping struct {
	WindowProfile WindowProfileCfg      `json:"window_profiles"`
	AxisDeadzone  int16                 `json:"axes_deadzone,omitempty"`
	Axes          map[uint8]AxisMapping `json:"axes,omitempty"`
	Buttons       map[uint8]KeyMapping  `json:"buttons,omitempty"`
	Hats          map[uint8]HatMapping  `json:"hats,omitempty"`
}

func (m *Mapping) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return nil
}

func (m *Mapping) WriteToFile(path string) error {
	data, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mapping) UpdateBinding(keyType, subKey string, index uint8, mode KeyMode, code int) {
	key := KeyMapping{Code: code, Mode: mode}

	switch keyType {
	case "button":
		if m.Buttons == nil {
			m.Buttons = make(map[uint8]KeyMapping)
		}
		m.Buttons[index] = key

	case "axis":
		if m.Axes == nil {
			m.Axes = make(map[uint8]AxisMapping)
		}
		axis := m.Axes[index]
		switch subKey {
		case "negative":
			axis.NegativeKey = key
		case "positive":
			axis.PositiveKey = key
		}
		m.Axes[index] = axis

	case "hat":
		if m.Hats == nil {
			m.Hats = make(map[uint8]HatMapping)
		}
		hat := m.Hats[index]
		switch subKey {
		case "up":
			hat.Up = key
		case "down":
			hat.Down = key
		case "left":
			hat.Left = key
		case "right":
			hat.Right = key
		}
		m.Hats[index] = hat
	}
}

func (m *Mapping) ClearBindings() {
	key := KeyMapping{Code: 0, Mode: 0}
	for k := range m.Axes {
		axis := m.Axes[k]
		axis.NegativeKey = key
		axis.PositiveKey = key
		m.Axes[k] = axis
	}
	for k := range m.Buttons {
		m.Buttons[k] = key
	}
	for k := range m.Hats {
		hat := m.Hats[k]
		hat.Up = key
		hat.Down = key
		hat.Left = key
		hat.Right = key
		m.Hats[k] = hat
	}
}
