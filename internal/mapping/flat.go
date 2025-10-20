package mapping

import (
	"fmt"
)

type FlatMapping struct {
	ButtonMap    map[uint8][]KeyMapping
	AxisPos      map[uint8][]KeyMapping
	AxisNeg      map[uint8][]KeyMapping
	AxisDeadzone int16
	HatDir       map[string][]KeyMapping
}

func (m *FlatMapping) Resolve(s *Store, evt JoystickEvent) ([]KeyMapping, []KeyMapping) {
	var pressed, released []KeyMapping

	markPressed := func(keys []KeyMapping) {
		for _, k := range keys {
			if !s.pressedKeys[k] {
				s.pressedKeys[k] = true
				pressed = append(pressed, k)
			}
		}
	}

	markReleased := func(keys []KeyMapping) {
		for _, k := range keys {
			if s.pressedKeys[k] {
				delete(s.pressedKeys, k)
				released = append(released, k)
			}
		}
	}

	switch evt.Type {
	case "button":
		if keys, ok := m.ButtonMap[evt.Index]; ok {
			if evt.Value > 0 {
				markPressed(keys)
			} else {
				markReleased(keys)
			}
		}

	case "hat":
		prev := s.lastHat[evt.Index]
		prevKey := key(evt.Index, dirVal(prev))
		curr := evt.Value
		currKey := key(evt.Index, dirVal(curr))

		if prevKey != currKey {
			if hatKeys, ok := m.HatDir[prevKey]; ok {
				markReleased(hatKeys)
			}
			if hatKeys, ok := m.HatDir[currKey]; ok {
				markPressed(hatKeys)
			}
		}

		s.lastHat[evt.Index] = curr

	case "axis":
		if keys, ok := m.ResolveAxisKey(evt.Index, evt.Value); ok {
			prev := s.lastAxis[evt.Index]
			var dir int8
			if evt.Value <= -m.AxisDeadzone {
				dir = -1
			} else if evt.Value >= m.AxisDeadzone {
				dir = +1
			}

			if prev != dir {
				if dir != 0 {
					markPressed(keys)
				} else {
					markReleased(keys)
				}
			}

			s.lastAxis[evt.Index] = dir
		}
	}

	return pressed, released
}

func (m *FlatMapping) ResolveAxisKey(axis uint8, value int16) ([]KeyMapping, bool) {
	if value > 0 {
		key, ok := m.AxisPos[axis]
		return key, ok
	} else if value < 0 {
		key, ok := m.AxisNeg[axis]
		return key, ok
	}
	return []KeyMapping{}, false
}

func (m *FlatMapping) GetKeys(keyType, subKey string, index uint8) []KeyMapping {
	var existingKeys []KeyMapping
	switch keyType {
	case "axis":
		switch subKey {
		case "positive":
			existingKeys = m.AxisPos[index]
		case "negative":
			existingKeys = m.AxisNeg[index]
		}
	case "hat":
		existingKeys = m.HatDir[key(index, subKey)]
	case "button":
		existingKeys = m.ButtonMap[index]
	}

	return existingKeys
}

func CompileFlatMapping(m Mapping) *FlatMapping {
	f := &FlatMapping{
		ButtonMap:    make(map[uint8][]KeyMapping),
		AxisPos:      make(map[uint8][]KeyMapping),
		AxisNeg:      make(map[uint8][]KeyMapping),
		HatDir:       make(map[string][]KeyMapping),
		AxisDeadzone: m.AxisDeadzone,
	}
	for k, v := range m.Buttons {
		f.ButtonMap[k] = v
	}
	for k, v := range m.Axes {
		f.AxisPos[k] = v.PositiveKey
		f.AxisNeg[k] = v.NegativeKey
	}
	for k, v := range m.Hats {
		f.HatDir[key(k, "up")] = v.Up
		f.HatDir[key(k, "down")] = v.Down
		f.HatDir[key(k, "left")] = v.Left
		f.HatDir[key(k, "right")] = v.Right
	}
	return f
}

func key(i uint8, dir string) string { return fmt.Sprintf("%d_%s", i, dir) }
func dirVal(i int16) string {
	switch i {
	case 1:
		return "up"
	case 2:
		return "right"
	case 4:
		return "down"
	case 8:
		return "left"
	default:
		return ""
	}
}
