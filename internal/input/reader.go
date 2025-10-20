package input

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/caedis/noreza/internal/mapping"
	"github.com/holoplot/go-evdev"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Reader struct {
	dev *evdev.InputDevice
}

func NewReader(path string) (*Reader, error) {
	dev, err := evdev.Open(path)
	if err != nil {
		return nil, err
	}
	return &Reader{dev: dev}, nil
}

func (r *Reader) Close() {
	r.dev.Close()
}

func (r *Reader) Stream(out chan<- mapping.JoystickEvent) {
	keyMap := make(map[evdev.EvCode]uint8)
	keyEvents := r.dev.CapableEvents(evdev.EV_KEY)
	for i, t := range keyEvents {
		keyMap[t] = uint8(i)
	}

	absInfos, err := r.dev.AbsInfos()
	if err != nil {
		log.Fatalln(err)
	}

	for {
		evt, err := r.dev.ReadOne()
		if err != nil {
			out <- mapping.JoystickEvent{}
			return
		}

		switch evt.Type {
		case evdev.EV_KEY:
			out <- mapping.JoystickEvent{Type: "button", Index: keyMap[evt.Code], Value: int16(evt.Value), Ready: true}
		case evdev.EV_ABS:
			switch evt.Code {
			case evdev.ABS_HAT0X:
				fallthrough
			case evdev.ABS_HAT0Y:
				var value int16
				if evt.Code == evdev.ABS_HAT0X {
					switch evt.Value {
					case -1:
						value = 8
					case 1:
						value = 2
					}
				} else {
					switch evt.Value {
					case -1:
						value = 1
					case 1:
						value = 4
					}
				}
				out <- mapping.JoystickEvent{Type: "hat", Index: 0, Value: value, Ready: true}
			default:
				if absInfo, ok := absInfos[evt.Code]; ok {
					scaled := scaleAxisToInt16(evt.Value, absInfo.Minimum, absInfo.Maximum)
					out <- mapping.JoystickEvent{Type: "axis", Index: uint8(evt.Code), Value: scaled, Ready: true}
				}
			}
		}

	}
}

func (r *Reader) String() string {
	if r.dev != nil {
		id, err := r.dev.InputID()
		if err != nil {
			return ""
		}
		product, err := mapping.GetDeviceFromID(id.Product)
		if err != nil {
			return ""
		}
		return cases.Title(language.English).String(product)
	}

	return ""
}

func GetDevicePath(serial string) (string, uint16, error) {
	basePath := "/dev/input/by-id"

	files, err := os.ReadDir(basePath)
	if err != nil {
		return "", 0, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.Contains(file.Name(), serial) && strings.HasSuffix(file.Name(), "event-joystick") {
			fullPath := filepath.Join(basePath, file.Name())
			dev, err := evdev.Open(fullPath)
			if err != nil {
				log.Printf("failed to open input device '%s': %v", file.Name(), err)
				continue
			}
			defer dev.Close()
			uid, _ := dev.UniqueID()
			if uid == serial {
				inputID, _ := dev.InputID()
				return fullPath, inputID.Product, nil
			}
		}
	}

	return "", 0, fmt.Errorf("device not found for serial '%s'", serial)
}

func scaleAxisToInt16(value int32, min int32, max int32) int16 {
	if max == min {
		return 0
	}
	scaled := (int64(value-min) * 65535 / int64(max-min)) - 32768
	return int16(scaled)
}
