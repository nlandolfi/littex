package login

import (
	"github.com/nlandolfi/lit/web-client/components/icons"
	"github.com/spinsrv/browser"
	"github.com/spinsrv/browser/dom"
	"github.com/spinsrv/browser/ui"
)

type State struct {
	Theme  *ui.Theme `json:"-"`
	Status *string

	Username, Password string
}

func (s *State) Handle(e browser.Event) {
	switch e.(type) {
	default:
	}
}

type EventLoginButtonClicked struct{}

func View(s *State) *browser.Node {
	return ui.VStack(
		ui.HStack(
			s.Theme.Card(
				ui.VStack(
					ui.HStack(
						icons.Trademark(s.Theme.BackgroundColor, 80),
					).JustifyContentCenter(),
					ui.VSpace(browser.Size{Value: 18, Unit: browser.UnitPX}),
					s.Theme.TextInput(&s.Username).
						FontSizePX(18).
						MarginPX(10).
						Placeholder("public").
						OnKeyDown(func(e dom.Event) {
							if e.KeyCode() == 13 { // enter
								go browser.Dispatch(EventLoginButtonClicked{}) // todo change name of action
							}
						}),
					s.Theme.PassInput(&s.Password).
						FontSizePX(18).
						MarginPX(10).
						Placeholder("private").
						OnKeyDown(func(e dom.Event) {
							if e.KeyCode() == 13 { // enter
								go browser.Dispatch(EventLoginButtonClicked{}) // todo change name of action
							}
						}),
					s.Theme.Button("Login").
						MarginPX(10).
						OnClickDispatch(EventLoginButtonClicked{}),
					ui.VStack(s.Theme.Text(*s.Status)).AlignItemsCenter(),
				).PaddingPX(20),
			).PaddingPX(30).
				MaxWidthPX(500),
		).JustifyContentCenter(),
	).JustifyContentCenter().
		HeightVH(100)
}
