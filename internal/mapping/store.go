package mapping

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type Profile struct {
	Name     string
	Active   bool
	Selected bool
}

type WindowProfile struct {
	NameRegex  *regexp.Regexp
	ClassRegex *regexp.Regexp
	Profile    string
}

type Store struct {
	ActiveMapping  atomic.Pointer[FlatMapping]
	Mappings       atomic.Pointer[map[string]*FlatMapping]
	RawMappings    atomic.Pointer[map[string]*Mapping]
	WindowProfiles atomic.Pointer[[]WindowProfile]
	// name of active profile
	ActiveProfile atomic.Value

	ProfilePath string
	ProductID   uint16
	// path to active symlink
	activePath  string
	lastHat     map[uint8]int16
	lastAxis    map[uint8]int8
	pressedKeys map[KeyMapping]bool
	eventSubs   atomic.Pointer[map[*chan SSEEvent]struct{}]
}

func NewStore(profilesPath string, productID uint16) *Store {
	s := Store{
		ProfilePath: profilesPath,
		activePath:  filepath.Join(profilesPath, "active"),
		ProductID:   productID,
		lastHat:     make(map[uint8]int16),
		lastAxis:    make(map[uint8]int8),
		pressedKeys: make(map[KeyMapping]bool),
	}

	s.eventSubs.Store(&map[*chan SSEEvent]struct{}{})

	return &s
}

func (s *Store) ListProfiles() []Profile {
	var out []Profile

	mappings := s.Mappings.Load()
	active := s.ActiveProfile.Load()
	for name := range *mappings {
		prof := Profile{
			Name:   name,
			Active: name == active,
		}
		out = append(out, prof)
	}

	slices.SortFunc(out, func(a, b Profile) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	return out
}

var unmarshallError *json.UnmarshalTypeError

func (s *Store) ReloadAllProfiles() error {
	files, err := os.ReadDir(s.ProfilePath)
	if err != nil {
		return err
	}

	mappings := make(map[string]*FlatMapping)
	rawMappings := make(map[string]*Mapping)
	var windowProfiles []WindowProfile

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.ProfilePath, f.Name()))
		if err != nil {
			continue
		}

		var m Mapping
		err = json.Unmarshal(data, &m)
		if errors.As(err, &unmarshallError) {
			s.migrateProfile(m, data)
			data, _ = json.MarshalIndent(m, "", "\t")
			go os.WriteFile(filepath.Join(s.ProfilePath, f.Name()), data, 0755)
		} else if err != nil {
			continue
		}

		name := strings.TrimSuffix(f.Name(), ".json")
		rawMappings[name] = &m
		mappings[name] = CompileFlatMapping(m)

		var compiled WindowProfile
		if m.WindowProfile.NamePattern != "" {
			compiled.NameRegex = regexp.MustCompile(m.WindowProfile.NamePattern)
		}
		if m.WindowProfile.ClassPattern != "" {
			compiled.ClassRegex = regexp.MustCompile(m.WindowProfile.ClassPattern)
		}
		if compiled.NameRegex != nil || compiled.ClassRegex != nil {
			compiled.Profile = name
			windowProfiles = append(windowProfiles, compiled)
		}

	}

	s.Mappings.Store(&mappings)
	s.RawMappings.Store(&rawMappings)
	s.WindowProfiles.Store(&windowProfiles)

	return nil
}

func (s *Store) ReloadProfile(name string) error {
	profileFile := filepath.Join(s.ProfilePath, name+".json")

	data, err := os.ReadFile(profileFile)
	if err != nil {
		return fmt.Errorf("read profile: %w", err)
	}

	var m Mapping
	err = json.Unmarshal(data, &m)
	if errors.As(err, &unmarshallError) {
		s.migrateProfile(m, data)
		data, _ = json.MarshalIndent(m, "", "\t")
		go os.WriteFile(profileFile, data, 0755)

	} else if err != nil {
		return fmt.Errorf("unmarshal profile: %w", err)
	}

	flat := CompileFlatMapping(m)

	mappingsPtr := s.Mappings.Load()
	rawMappingsPtr := s.RawMappings.Load()
	windowProfilesPtr := s.WindowProfiles.Load()

	mappings := make(map[string]*FlatMapping)
	rawMappings := make(map[string]*Mapping)
	windowProfiles := make([]WindowProfile, 0)

	if mappingsPtr != nil {
		for k, v := range *mappingsPtr {
			mappings[k] = v
		}
	}
	if rawMappingsPtr != nil {
		for k, v := range *rawMappingsPtr {
			rawMappings[k] = v
		}
	}
	if windowProfilesPtr != nil {
		windowProfiles = append(windowProfiles, *windowProfilesPtr...)
	}

	mappings[name] = flat
	rawMappings[name] = &m

	// Remove any previous window matchers that pointed to this profile
	filtered := windowProfiles[:0]
	for _, wp := range windowProfiles {
		if wp.Profile != name {
			filtered = append(filtered, wp)
		}
	}
	windowProfiles = filtered

	// Add any new matchers defined in this profile
	def := m.WindowProfile
	var compiled WindowProfile
	if def.NamePattern != "" {
		if r, err := regexp.Compile(def.NamePattern); err == nil {
			compiled.NameRegex = r
		} else {
			fmt.Printf("invalid name regex for %s: %v\n", name, err)
		}
	}
	if def.ClassPattern != "" {
		if r, err := regexp.Compile(def.ClassPattern); err == nil {
			compiled.ClassRegex = r
		} else {
			fmt.Printf("invalid class regex for %s: %v\n", name, err)
		}
	}
	if compiled.NameRegex != nil || compiled.ClassRegex != nil {
		compiled.Profile = name
		windowProfiles = append(windowProfiles, compiled)
	}

	// Atomically store back
	s.Mappings.Store(&mappings)
	s.RawMappings.Store(&rawMappings)
	s.WindowProfiles.Store(&windowProfiles)

	return nil
}

func (s *Store) RemoveProfile(name string) {
	mappingsPtr := s.Mappings.Load()
	rawMappingsPtr := s.RawMappings.Load()
	windowProfilesPtr := s.WindowProfiles.Load()

	if mappingsPtr == nil || rawMappingsPtr == nil {
		return
	}

	mappings := make(map[string]*FlatMapping)
	rawMappings := make(map[string]*Mapping)
	windowProfiles := make([]WindowProfile, 0)

	for k, v := range *mappingsPtr {
		if k != name {
			mappings[k] = v
		}
	}
	for k, v := range *rawMappingsPtr {
		if k != name {
			rawMappings[k] = v
		}
	}
	if windowProfilesPtr != nil {
		for _, wp := range *windowProfilesPtr {
			if wp.Profile != name {
				windowProfiles = append(windowProfiles, wp)
			}
		}
	}

	log.Println("removing profile:", name)

	s.Mappings.Store(&mappings)
	s.RawMappings.Store(&rawMappings)
	s.WindowProfiles.Store(&windowProfiles)
}

func (s *Store) WatchProfiles(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(s.ProfilePath); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-watcher.Events:
			if !ok {
				continue
			}
			base := filepath.Base(ev.Name)

			// React when the active symlink changes
			if base == "active" &&
				(ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename)) != 0 {
				if err := s.ReloadActive(); err != nil {
					fmt.Println("reload error:", err)
				}
				continue
			}

			if filepath.Ext(base) == ".json" {
				name := strings.TrimSuffix(base, ".json")
				if ev.Op.Has(fsnotify.Write) {
					if err := s.ReloadProfile(name); err != nil {
						log.Println("WatchProfiles:", err)
					}
				} else if ev.Op.Has(fsnotify.Remove) {
					s.RemoveProfile(name)
				}
			}

		case err := <-watcher.Errors:
			fmt.Println("watcher error:", err)
		}
	}
}

func (s *Store) CreateIfNeeded() error {
	activePath := filepath.Join(s.ProfilePath, "active")
	defaultPath := filepath.Join(s.ProfilePath, "default.json")
	if _, err := os.Lstat(activePath); os.IsNotExist(err) {
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			if err := s.CreateNewProfile("default.json"); err != nil {
				return err
			}
		}

		// relative on purpose
		err := os.Symlink("default.json", activePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateNewProfile(fileName string) error {
	targetPath := filepath.Join(s.ProfilePath, fileName)

	if _, err := os.Stat(targetPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	defaultBytes, err := s.getDefaultMapping()
	if err != nil {
		return err
	}
	if err = os.WriteFile(targetPath, defaultBytes, 0755); err != nil {
		return err
	}

	s.ReloadProfile(strings.TrimSuffix(fileName, ".json"))

	return nil
}

func (s *Store) SetActiveProfile(name string) error {
	target := filepath.Join(s.ProfilePath, name+".json")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Errorf("profile %s not found", name)
	}

	tmpPath := s.activePath + ".tmp"
	_ = os.Remove(tmpPath)

	if err := os.Symlink(name+".json", tmpPath); err != nil {
		return fmt.Errorf("failed to create tmp symlink: %w", err)
	}

	if err := os.Rename(tmpPath, s.activePath); err != nil {
		return fmt.Errorf("failed to swap active symlink: %w", err)
	}

	// Set active in memory
	s.setActive(name)

	s.BroadcastEvent(SSEEvent{
		Type: EventActiveProfile,
		Data: name,
	})

	return nil
}

func (s *Store) ReleaseAll() []KeyMapping {
	if len(s.pressedKeys) == 0 {
		return nil
	}

	released := make([]KeyMapping, 0, len(s.pressedKeys))
	for k := range s.pressedKeys {
		released = append(released, k)
	}
	s.pressedKeys = make(map[KeyMapping]bool)
	return released
}

func (s *Store) setActive(name string) {
	s.ActiveProfile.Store(name)

	mappingsPtr := s.Mappings.Load()
	if mappingsPtr == nil {
		s.ActiveMapping.Store(nil)
		return
	}

	if flat, ok := (*mappingsPtr)[name]; ok {
		s.ActiveMapping.Store(flat)
	} else {
		s.ActiveMapping.Store(nil)
	}
}

func (s *Store) migrateProfile(newMapping Mapping, oldData []byte) error {
	var oldMapping Mapping_V1
	if err := json.Unmarshal(oldData, &oldMapping); err != nil {
		return fmt.Errorf("Error migrating profile")
	}

	for i, axe := range oldMapping.Axes {
		newMapping.Axes[i] = AxisMapping{
			PositiveKey: []KeyMapping{
				{Code: axe.PositiveKey.Code, Mode: KeyMode(axe.PositiveKey.Mode)},
			},
			NegativeKey: []KeyMapping{
				{Code: axe.NegativeKey.Code, Mode: KeyMode(axe.NegativeKey.Mode)},
			},
		}
	}

	for i, hat := range oldMapping.Hats {
		newMapping.Hats[i] = HatMapping{
			Up: []KeyMapping{
				{Code: hat.Up.Code, Mode: KeyMode(hat.Up.Mode)},
			},
			Down: []KeyMapping{
				{Code: hat.Down.Code, Mode: KeyMode(hat.Down.Mode)},
			},
			Left: []KeyMapping{
				{Code: hat.Left.Code, Mode: KeyMode(hat.Left.Mode)},
			},
			Right: []KeyMapping{
				{Code: hat.Right.Code, Mode: KeyMode(hat.Right.Mode)},
			},
		}
	}

	for i, key := range oldMapping.Buttons {
		newMapping.Buttons[i] = []KeyMapping{
			{Code: key.Code, Mode: KeyMode(key.Mode)},
		}
	}

	return nil
}

func (s *Store) ReloadActive() error {
	target, err := os.Readlink(s.activePath)
	if err != nil {
		return fmt.Errorf("readlink: %w", err)
	}

	name := strings.TrimSuffix(filepath.Base(target), ".json")

	if err := s.ReloadProfile(name); err != nil {
		return err
	}

	s.setActive(name)
	return nil
}

func (s *Store) Resolve(evt JoystickEvent) ([]KeyMapping, []KeyMapping) {
	m := s.ActiveMapping.Load()
	if m == nil {
		return nil, nil
	}
	return m.Resolve(s, evt)
}

func GetDeviceFromID(productID uint16) (string, error) {
	switch productID {
	case 3903:
		return "classic", nil
	case 4284:
		return "cyborg", nil
	case 4355:
		return "cryo", nil
	case 4412:
		return "cyborg-tansy", nil
	case 4498:
		return "classic-tansy", nil
	case 4626:
		return "cryo-lefty", nil
	case 4855:
		return "cyborg2", nil
	case 5098:
		return "keyzen", nil
	}

	return "", fmt.Errorf("product id '%x' does not match a valid device", productID)
}

func (s *Store) SaveProfile(name string) error {
	raw := s.RawMappings.Load()
	m, ok := (*raw)[name]
	if !ok {
		return fmt.Errorf("profile not found: %s", name)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	go os.WriteFile(filepath.Join(s.ProfilePath, name+".json"), data, 0755)
	return nil
}

type EventType string

const (
	EventJoystick        EventType = "joystick"
	EventActiveProfile   EventType = "activeProfile"
	EventSelectedProfile EventType = "selectedProfile"
)

type SSEEvent struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}

func (s *Store) BroadcastEvent(e SSEEvent) {
	subs := *s.eventSubs.Load()
	for ch := range subs {
		select {
		case *ch <- e:
		default:
			// Drop if slow
		}
	}
}

func (s *Store) Subscribe() *chan SSEEvent {
	ch := make(chan SSEEvent, 32)
	old := *s.eventSubs.Load()
	newMap := make(map[*chan SSEEvent]struct{}, len(old)+1)
	for k := range old {
		newMap[k] = struct{}{}
	}
	newMap[&ch] = struct{}{}
	s.eventSubs.Store(&newMap)
	return &ch
}

func (s *Store) Unsubscribe(ch *chan SSEEvent) {
	old := *s.eventSubs.Load()
	newMap := make(map[*chan SSEEvent]struct{}, len(old))
	for k := range old {
		if k != ch {
			newMap[k] = struct{}{}
		}
	}
	s.eventSubs.Store(&newMap)
}

//go:embed default_mappings/*
var defaultMappings embed.FS

func (s *Store) getDefaultMapping() ([]byte, error) {
	device, err := GetDeviceFromID(s.ProductID)
	if err != nil {
		return nil, err
	}

	return defaultMappings.ReadFile(fmt.Sprintf("default_mappings/%s.json", device))
}
