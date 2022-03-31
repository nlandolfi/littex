package profile

import (
	"fmt"

	"github.com/spinsrv/browser"
	"github.com/spinsrv/browser/ui"
	"github.com/spinsrv/spin"
)

type State struct {
	Theme *ui.Theme `json:"-"`

	PrivateKey **spin.PrivateKey `json:"-"`
}

func (s *State) Handle(e browser.Event) {
	switch e.(type) {
	default:
	}
}

type EventSelectTheme struct{ Key string }

func View(s *State) *browser.Node {
	return ui.VStack(
		s.Theme.Textf("Hello %s!", (*s.PrivateKey).Citizen),
		s.Theme.Link(fmt.Sprintf("https://serve.spinsrv.com/%s/", (*s.PrivateKey).Citizen), "Serve"),
		s.Theme.Link("https://keys.spinsrv.com", "Keys"),
		s.Theme.Link("https://dir.spinsrv.com", "Dir"),
		s.Theme.Link("https://store.spinsrv.com", "Store"),
		s.Theme.Button("Light Theme").OnClickDispatch(EventSelectTheme{"light"}),
		s.Theme.Button("Dark Theme").OnClickDispatch(EventSelectTheme{"dark"}),
		s.Theme.Button("Default Theme").OnClickDispatch(EventSelectTheme{"default"}),
	)
}
