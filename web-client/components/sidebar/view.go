package sidebar

import (
	"fmt"
	"strings"

	"github.com/spinsrv/browser"
	"github.com/spinsrv/browser/ui"
)

type State struct {
	Theme  *ui.Theme `json:"-"`
	Prefix string

	hoveredKey      string
	SelectedKey     string
	SelectedDisplay string
}

func (s *State) Handle(e browser.Event) {
	switch e := e.(type) {
	case EventItemClick:
		if e.Target != s {
			return
		}
		s.SelectedKey = e.Key
		s.SelectedDisplay = e.Display
	case EventItemHoverStart:
		if e.Target != s {
			return
		}
		s.hoveredKey = e.Key
	case EventItemHoverEnd:
		if e.Target != s {
			return
		}
		if s.hoveredKey == e.Key {
			s.hoveredKey = ""
		}
	default:
	}
}

type Item struct {
	Key     string
	Display string
}

// structs for extensibility
type EventItemHoverStart struct {
	Target *State
	Item
}
type EventItemHoverEnd struct {
	Target *State
	Item
}
type EventItemClick struct {
	Target *State
	Item
}

func View(s *State, items []*Item) *browser.Node {
	views := make([]*browser.Node, len(items))

	for index, item := range items {
		views[index] = itemView(s, index, item)
	}

	return ui.VStack(views...)
}

func itemView(s *State, index int, item *Item) *browser.Node {
	return ui.Div(
		ui.If(
			item.Key == s.SelectedKey,
			func() *browser.Node {
				return ui.HStack(
					s.Theme.Text(item.Display),
					s.Theme.Text("›").MarginLeftPX(10),
				).JustifyContentSpaceBetween()
			},
			func() *browser.Node {
				return ui.If(
					item.Key == s.hoveredKey,
					func() *browser.Node {
						return ui.HStack(
							s.Theme.Text(item.Display),
							ui.If(strings.Contains(item.Key, "out"),
								func() *browser.Node {
									return s.Theme.Text("‹").MarginLeftPX(10)
								},
								func() *browser.Node {
									return s.Theme.Text("›").MarginLeftPX(10)
								},
							),
						).JustifyContentSpaceBetween()
					},
					func() *browser.Node {
						return s.Theme.Text(item.Display) // TODO, add the arrow.
					},
				)
			},
		),
	).
		PaddingPX(9).
		BorderRadiusPX(5).
		CursorPointer().
		OnlyIf(s.hoveredKey == item.Key, func(n *browser.Node) *browser.Node {
			return n.Background(s.Theme.HoverBackgroundColor)
		}).
		OnlyIf(s.SelectedKey == item.Key || (s.SelectedKey == "" && index == 0), func(n *browser.Node) *browser.Node {
			return n.Color("blue")
		}).
		OnClickDispatch(EventItemClick{s, *item}).
		OnMouseEnterDispatch(EventItemHoverStart{s, *item}).
		OnMouseLeaveDispatch(EventItemHoverEnd{s, *item}).
		ID(fmt.Sprintf("sidebar-%s", item.Key))
}
