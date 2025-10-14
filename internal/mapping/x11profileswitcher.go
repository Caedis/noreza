package mapping

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type AutoProfileSwitcher struct {
	conn       *xgb.Conn
	store      *Store
	pollPeriod time.Duration
}

func NewAutoProfileSwitcher(store *Store, pollPeriod time.Duration) (*AutoProfileSwitcher, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}

	return &AutoProfileSwitcher{
		conn:       conn,
		store:      store,
		pollPeriod: pollPeriod,
	}, nil
}

// Runs the polling loop (blocking)
func (a *AutoProfileSwitcher) Start(ctx context.Context) {

	setup := xproto.Setup(a.conn)
	screen := setup.DefaultScreen(a.conn)
	rootWin := screen.Root

	ticker := time.NewTicker(a.pollPeriod)
	defer ticker.Stop()

	var lastWin xproto.Window
	for {
		select {
		case <-ticker.C:
			win, err := a.getFocusedWindow()
			if err != nil {
				log.Println("Error getting focused window:", err)
				continue
			}
			if win == rootWin {
				continue
			}

			topMostWin, err := a.getTopLevelWindow(win, rootWin)
			if err != nil {
				log.Println("Error getting topmost window:", err)
				continue
			}

			if topMostWin != rootWin {
				win = topMostWin
			}

			if win != lastWin {
				lastWin = win
				a.switchProfileForWindow(win)
			}
		case <-ctx.Done():
			return
		}
	}

}

// Gets the input focused window.
// More consistant than _NET_ACTIVE_WINDOW
func (a *AutoProfileSwitcher) getFocusedWindow() (xproto.Window, error) {
	reply, err := xproto.GetInputFocus(a.conn).Reply()
	if err != nil {
		return 0, err
	}
	return reply.Focus, nil
}

// Checks window properties and updates the active profile.
func (a *AutoProfileSwitcher) switchProfileForWindow(win xproto.Window) {
	store := a.store

	name, err := a.getWindowName(win)
	if err != nil {
		log.Printf("switchProfileForWindow: %s\n", err)
		return
	}
	class, err := a.getWindowClass(win)
	if err != nil {
		log.Printf("switchProfileForWindow: %s\n", err)
		return
	}

	windowProfiles := *store.WindowProfiles.Load()
	for _, wp := range windowProfiles {
		match := false

		if wp.NameRegex != nil && wp.NameRegex.MatchString(name) {
			match = true
		}

		if wp.ClassRegex != nil && wp.ClassRegex.MatchString(class) {
			match = true
		}

		if match {
			store.SetActiveProfile(wp.Profile)
			return
		}
	}
}

func (a *AutoProfileSwitcher) getWindowName(win xproto.Window) (string, error) {
	prop, err := xproto.GetProperty(a.conn, false, win, xproto.AtomWmName, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil || prop == nil {
		return "", err
	}
	return string(prop.Value), nil
}

func (a *AutoProfileSwitcher) getWindowClass(win xproto.Window) (string, error) {
	prop, err := xproto.GetProperty(a.conn, false, win, xproto.AtomWmClass, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil || prop == nil {
		return "", err
	}
	parts := bytes.Split(prop.Value, []byte{0})

	switch len(parts) {
	case 3:
		fallthrough
	case 2:
		return string(parts[1]), nil
	case 1:
		return string(parts[0]), nil
	}

	return "", nil
}

func (a *AutoProfileSwitcher) getTopLevelWindow(win xproto.Window, root xproto.Window) (xproto.Window, error) {
	for {
		tree, err := xproto.QueryTree(a.conn, win).Reply()
		if err != nil {
			return 0, err
		}
		if tree.Parent == root {
			return win, nil
		}
		win = tree.Parent
	}
}
