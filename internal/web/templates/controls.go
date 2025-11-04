package templates

import (
	"github.com/a-h/templ"
	"github.com/caedis/noreza/internal/mapping"
)

type Control interface {
	Position(totalCols int, opposite bool) (row, col int)
	Render(m mapping.Mapping, profile string, row, col int) templ.Component
}

type ButtonControl struct {
	Row   int
	Col   int
	Index uint8
}

// Compute mirrored or normal position
func (b ButtonControl) Position(totalCols int, opposite bool) (int, int) {
	col := b.Col
	if opposite {
		col = totalCols - col + 1
	}
	return b.Row, col
}

func (c ButtonControl) Render(m mapping.Mapping, profile string, row, col int) templ.Component {
	return MappingButton(c.Index, m, row, col, profile)
}

type HatControl struct {
	ButtonControl
	CenterButtonIndex uint8
}

func (c HatControl) Position(totalCols int, opposite bool) (int, int) {
	col := c.Col
	if opposite {
		col = totalCols - col - 1
	}
	return c.Row, col
}

func (c HatControl) Render(m mapping.Mapping, profile string, row, col int) templ.Component {
	return Hat(c.Index, m, row, col, profile, c.CenterButtonIndex)
}

type JoystickControl struct {
	ButtonControl
	YIndex            uint8
	CenterButtonIndex uint8
}

func (c JoystickControl) Render(m mapping.Mapping, profile string, row, col int) templ.Component {
	return Joystick(c.Index, c.YIndex, m, row, col, profile, c.CenterButtonIndex)
}

func (c JoystickControl) Position(totalCols int, opposite bool) (int, int) {
	col := c.Col
	if opposite {
		col = totalCols - col - 1
	}
	return c.Row, col
}

type CyroScrollControl struct {
	ButtonControl
}

func (c CyroScrollControl) Render(m mapping.Mapping, profile string, row, col int) templ.Component {
	return CyroScroll(c.Index, m, row, col, profile)
}

var ClassicLayout = []Control{
	HatControl{ButtonControl: ButtonControl{Row: 1, Col: 7, Index: 0}, CenterButtonIndex: 21},
	JoystickControl{ButtonControl: ButtonControl{Row: 4, Col: 7, Index: 0}, YIndex: 1, CenterButtonIndex: 22},
	ButtonControl{Row: 1, Col: 3, Index: 1},
	ButtonControl{Row: 1, Col: 4, Index: 2},
	ButtonControl{Row: 2, Col: 1, Index: 3},
	ButtonControl{Row: 2, Col: 2, Index: 4},
	ButtonControl{Row: 2, Col: 3, Index: 5},
	ButtonControl{Row: 2, Col: 4, Index: 6},
	ButtonControl{Row: 3, Col: 1, Index: 7},
	ButtonControl{Row: 3, Col: 2, Index: 8},
	ButtonControl{Row: 3, Col: 3, Index: 9},
	ButtonControl{Row: 3, Col: 4, Index: 10},
	ButtonControl{Row: 4, Col: 1, Index: 11},
	ButtonControl{Row: 4, Col: 2, Index: 12},
	ButtonControl{Row: 4, Col: 3, Index: 13},
	ButtonControl{Row: 4, Col: 4, Index: 14},
	ButtonControl{Row: 4, Col: 5, Index: 15},
	ButtonControl{Row: 5, Col: 1, Index: 16},
	ButtonControl{Row: 5, Col: 2, Index: 17},
	ButtonControl{Row: 5, Col: 3, Index: 18},
	ButtonControl{Row: 5, Col: 4, Index: 19},
	ButtonControl{Row: 5, Col: 10, Index: 20},
}

var Cyborg1Layout = []Control{
	ButtonControl{Row: 1, Col: 2, Index: 1},
	ButtonControl{Row: 1, Col: 3, Index: 2},
	ButtonControl{Row: 1, Col: 4, Index: 3},
	ButtonControl{Row: 1, Col: 5, Index: 4},
	HatControl{ButtonControl: ButtonControl{Row: 1, Col: 7, Index: 0}, CenterButtonIndex: 25},
	ButtonControl{Row: 2, Col: 2, Index: 5},
	ButtonControl{Row: 2, Col: 3, Index: 6},
	ButtonControl{Row: 2, Col: 4, Index: 7},
	ButtonControl{Row: 2, Col: 5, Index: 8},
	ButtonControl{Row: 3, Col: 1, Index: 9},
	ButtonControl{Row: 3, Col: 2, Index: 10},
	ButtonControl{Row: 3, Col: 3, Index: 11},
	ButtonControl{Row: 3, Col: 4, Index: 12},
	ButtonControl{Row: 3, Col: 5, Index: 13},
	ButtonControl{Row: 3, Col: 6, Index: 14},
	JoystickControl{ButtonControl: ButtonControl{Row: 4, Col: 7, Index: 0}, YIndex: 1, CenterButtonIndex: 28},
	ButtonControl{Row: 4, Col: 2, Index: 15},
	ButtonControl{Row: 4, Col: 3, Index: 16},
	ButtonControl{Row: 4, Col: 4, Index: 17},
	ButtonControl{Row: 4, Col: 5, Index: 18},
	ButtonControl{Row: 5, Col: 2, Index: 19},
	ButtonControl{Row: 5, Col: 3, Index: 20},
	ButtonControl{Row: 5, Col: 4, Index: 21},
	ButtonControl{Row: 5, Col: 5, Index: 22},
	ButtonControl{Row: 5, Col: 10, Index: 23},
}

var Cyborg2Layout = []Control{
	ButtonControl{Row: 1, Col: 2, Index: 1},
	ButtonControl{Row: 1, Col: 3, Index: 2},
	ButtonControl{Row: 1, Col: 4, Index: 3},
	ButtonControl{Row: 1, Col: 5, Index: 4},
	HatControl{ButtonControl: ButtonControl{Row: 1, Col: 7, Index: 0}, CenterButtonIndex: 25},
	ButtonControl{Row: 2, Col: 2, Index: 5},
	ButtonControl{Row: 2, Col: 3, Index: 6},
	ButtonControl{Row: 2, Col: 4, Index: 7},
	ButtonControl{Row: 2, Col: 5, Index: 8},
	ButtonControl{Row: 3, Col: 1, Index: 9},
	ButtonControl{Row: 3, Col: 2, Index: 10},
	ButtonControl{Row: 3, Col: 3, Index: 11},
	ButtonControl{Row: 3, Col: 4, Index: 12},
	ButtonControl{Row: 3, Col: 5, Index: 13},
	ButtonControl{Row: 3, Col: 6, Index: 14},
	JoystickControl{ButtonControl: ButtonControl{Row: 4, Col: 7, Index: 0}, YIndex: 1, CenterButtonIndex: 28},
	ButtonControl{Row: 4, Col: 2, Index: 15},
	ButtonControl{Row: 4, Col: 3, Index: 16},
	ButtonControl{Row: 4, Col: 4, Index: 17},
	ButtonControl{Row: 4, Col: 5, Index: 18},
	ButtonControl{Row: 5, Col: 2, Index: 19},
	ButtonControl{Row: 5, Col: 3, Index: 20},
	ButtonControl{Row: 5, Col: 4, Index: 21},
	ButtonControl{Row: 5, Col: 5, Index: 22},
	ButtonControl{Row: 5, Col: 10, Index: 23},
	ButtonControl{Row: 6, Col: 10, Index: 24},
}

var CyroLayout = []Control{
	HatControl{ButtonControl: ButtonControl{Row: 1, Col: 1, Index: 0}, CenterButtonIndex: 17},
	JoystickControl{ButtonControl: ButtonControl{Row: 4, Col: 1, Index: 0}, YIndex: 1, CenterButtonIndex: 18},
	ButtonControl{Row: 3, Col: 4, Index: 1},
	ButtonControl{Row: 3, Col: 5, Index: 2},
	ButtonControl{Row: 3, Col: 6, Index: 3},
	ButtonControl{Row: 3, Col: 7, Index: 4},
	ButtonControl{Row: 4, Col: 4, Index: 5},
	ButtonControl{Row: 4, Col: 5, Index: 6},
	ButtonControl{Row: 4, Col: 6, Index: 7},
	ButtonControl{Row: 4, Col: 7, Index: 8},
	CyroScrollControl{ButtonControl{Row: 5, Col: 4, Index: 19}},
	ButtonControl{Row: 5, Col: 5, Index: 9},
	ButtonControl{Row: 5, Col: 6, Index: 10},
	ButtonControl{Row: 5, Col: 7, Index: 11},
	ButtonControl{Row: 5, Col: 8, Index: 12},
	ButtonControl{Row: 6, Col: 5, Index: 13},
	ButtonControl{Row: 6, Col: 6, Index: 14},
	ButtonControl{Row: 6, Col: 7, Index: 15},
	ButtonControl{Row: 6, Col: 8, Index: 16},
}

var KeyzenLayout = []Control{
	ButtonControl{Row: 1, Col: 3, Index: 1},
	ButtonControl{Row: 1, Col: 4, Index: 2},
	ButtonControl{Row: 1, Col: 5, Index: 3},
	HatControl{ButtonControl: ButtonControl{Row: 1, Col: 7, Index: 0}, CenterButtonIndex: 27},
	ButtonControl{Row: 2, Col: 2, Index: 4},
	ButtonControl{Row: 2, Col: 3, Index: 5},
	ButtonControl{Row: 2, Col: 4, Index: 6},
	ButtonControl{Row: 2, Col: 5, Index: 7},
	ButtonControl{Row: 2, Col: 6, Index: 8},
	ButtonControl{Row: 3, Col: 1, Index: 9},
	ButtonControl{Row: 3, Col: 2, Index: 10},
	ButtonControl{Row: 3, Col: 3, Index: 11},
	ButtonControl{Row: 3, Col: 4, Index: 12},
	ButtonControl{Row: 3, Col: 5, Index: 13},
	ButtonControl{Row: 3, Col: 6, Index: 14},
	JoystickControl{ButtonControl{Row: 4, Col: 8, Index: 0}, 1, 28},
	ButtonControl{Row: 4, Col: 1, Index: 15},
	ButtonControl{Row: 4, Col: 2, Index: 16},
	ButtonControl{Row: 4, Col: 3, Index: 17},
	ButtonControl{Row: 4, Col: 4, Index: 18},
	ButtonControl{Row: 4, Col: 5, Index: 19},
	ButtonControl{Row: 4, Col: 6, Index: 20},
	ButtonControl{Row: 5, Col: 2, Index: 21},
	ButtonControl{Row: 5, Col: 3, Index: 22},
	ButtonControl{Row: 5, Col: 4, Index: 23},
	ButtonControl{Row: 5, Col: 5, Index: 24},
	ButtonControl{Row: 5, Col: 11, Index: 25},
	ButtonControl{Row: 6, Col: 11, Index: 26},
}
