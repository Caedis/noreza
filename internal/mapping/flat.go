package mapping

import (
	"fmt"
)

type FlatMapping struct {
	ButtonMap    map[uint8]KeyMapping
	AxisPos      map[uint8]KeyMapping
	AxisNeg      map[uint8]KeyMapping
	AxisDeadzone int16
	HatDir       map[string]KeyMapping
}

func (m *FlatMapping) Resolve(s *Store, evt JoystickEvent) ([]KeyMapping, []KeyMapping) {
	var pressed, released []KeyMapping

	switch evt.Type {
	case "button":
		if key, ok := m.ButtonMap[evt.Index]; ok {
			if evt.Value > 0 {
				pressed = append(pressed, key)
			} else {
				released = append(released, key)
			}
		}
	case "hat":
		prev := s.lastHat[evt.Index]
		prevKey := key(evt.Index, dirVal(prev))

		curr := evt.Value
		currKey := key(evt.Index, dirVal(curr))

		if hat, ok := m.HatDir[prevKey]; ok {
			released = append(released, hat)
		}
		if hat, ok := m.HatDir[currKey]; ok {
			pressed = append(pressed, hat)
		}

		s.lastHat[evt.Index] = curr
	case "axis":
		if key, ok := m.ResolveAxisKey(evt.Index, evt.Value); ok {

			prev := s.lastAxis[evt.Index]
			var dir int8
			if evt.Value <= -m.AxisDeadzone {
				dir = -1
			} else if evt.Value >= m.AxisDeadzone {
				dir = +1
			}

			if prev != dir {
				if dir != 0 {
					pressed = append(pressed, key)
				} else {
					released = append(released, key)
				}
			}

			s.lastAxis[evt.Index] = dir

		}
	}

	return pressed, released
}

func (m *FlatMapping) ResolveAxisKey(axis uint8, value int16) (KeyMapping, bool) {
	if value > 0 {
		key, ok := m.AxisPos[axis]
		return key, ok
	} else if value < 0 {
		key, ok := m.AxisNeg[axis]
		return key, ok
	}
	return KeyMapping{}, false
}

func CompileFlatMapping(m Mapping) *FlatMapping {
	f := &FlatMapping{
		ButtonMap:    make(map[uint8]KeyMapping),
		AxisPos:      make(map[uint8]KeyMapping),
		AxisNeg:      make(map[uint8]KeyMapping),
		HatDir:       make(map[string]KeyMapping),
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
