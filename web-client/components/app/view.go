package app

import (
	"log"
	"time"

	"github.com/nlandolfi/lit/web-client/components/icons"
	"github.com/nlandolfi/lit/web-client/components/litterae"
	"github.com/nlandolfi/lit/web-client/components/login"
	"github.com/nlandolfi/lit/web-client/components/profile"
	"github.com/nlandolfi/lit/web-client/components/sidebar"
	"github.com/spinsrv/browser"
	"github.com/spinsrv/browser/ui"
	"github.com/spinsrv/spin"
	"github.com/spinsrv/spin/key"
)

type State struct {
	ui.Theme

	*spin.PrivateKey
	LoginError string
	key.KeyServerHTTPClient

	SidebarState  sidebar.State
	SidebarHidden bool

	LoginState    login.State
	LitteraeState litterae.State
	ProfileState  profile.State

	ClientVersion string
	LastWrittenAt time.Time
}

func (s *State) Handle(e browser.Event) {
	switch v := e.(type) {
	case EventInitialize:
		return
	case EventLoginSuccess:
		log.Print("success")
		s.LoginState.Username = ""
		s.LoginState.Password = ""
		go browser.Dispatch(sidebar.EventItemClick{Target: &s.SidebarState, Item: *items[0]})
	case EventToggleSidebar:
		s.SidebarHidden = !s.SidebarHidden
	case login.EventLoginButtonClicked:
		go s.Login(s.LoginState.Username, s.LoginState.Password)
	case profile.EventSelectTheme:
		s.SetTheme(v.Key)
	case sidebar.EventItemClick:
		switch v.Key {
		case "logout":
			s.Logout()
		default:
			s.SidebarState.Handle(e)
		}
	case sidebar.EventItemHoverStart, sidebar.EventItemHoverEnd:
		s.SidebarState.Handle(e)
	default:
	}

	s.LoginState.Handle(e)
	s.ProfileState.Handle(e)
	s.LitteraeState.Handle(e)
}

type EventInitialize struct{}
type EventLoginSuccess struct{}
type EventToggleSidebar struct{}

func (s *State) Login(pu, pr string) {
	s.LoginError = "authenticating..."
	go browser.Dispatch(nil)
	defer func() { go browser.Dispatch(nil) }()

	resp := s.KeyServerHTTPClient.Temp(&spin.KeyTempRequest{
		Public:   pu,
		Private:  pr,
		Duration: 24 * time.Hour,
	})

	if resp.Error != "" {
		s.LoginError = resp.Error
		return
	}

	if resp.Key == nil {
		s.LoginError = "authentication failed"
		return
	}

	if s.PrivateKey == nil {
		s.PrivateKey = new(spin.PrivateKey)
	}

	go browser.Dispatch(EventLoginSuccess{})

	s.PrivateKey.Key = *resp.Key
	s.PrivateKey.Private = resp.Private
	s.LoginError = ""
}

func (s *State) Logout() {
	t := s.Theme
	v := s.ClientVersion
	*s = State{} // wipe state
	s.Theme = t
	s.ClientVersion = v
	s.Rewire()
}

func (s *State) Rewire() {
	s.LoginState.Theme = &s.Theme
	s.LoginState.Status = &s.LoginError
	s.LitteraeState.PrivateKey = &s.PrivateKey
	s.LitteraeState.Theme = &s.Theme
	s.ProfileState.PrivateKey = &s.PrivateKey
	s.ProfileState.Theme = &s.Theme
	s.SidebarState.Theme = &s.Theme
}

func (s *State) SetTheme(to string) {
	switch to {
	case "light":
		s.Theme = LightTheme
	case "dark":
		s.Theme = DarkTheme
	default:
		s.Theme = DefaultTheme
	}
}

func View(s *State) *browser.Node {
	return view(s).Background(s.Theme.BackgroundColor).FontFamily(s.Theme.FontFamily)
}

var items = []*sidebar.Item{
	&sidebar.Item{
		Key:     "litterae",
		Display: "Litterae",
	},
	&sidebar.Item{
		Key:     "profile",
		Display: "Profile",
	},
	&sidebar.Item{
		Key:     "logout",
		Display: "Logout",
	},
}

func view(s *State) *browser.Node {
	if s.PrivateKey == nil {
		return login.View(&s.LoginState).Background("black")
	}

	var view *browser.Node
	switch s.SidebarState.SelectedKey {
	case "litterae", "":
		view = litterae.View(&s.LitteraeState)
	case "profile":
		view = profile.View(&s.ProfileState)
	default:
		log.Fatalf("unknown selected app: %q", s.SidebarState.SelectedKey)
	}

	return ui.VStack(
		header(s),
		ui.HStack(
			ui.OnlyIf(!s.SidebarHidden,
				func() *browser.Node {
					return sidebar.View(&s.SidebarState, items).
						Width(browser.Size{Value: 100, Unit: browser.UnitPX}).
						PaddingPX(10).
						BorderRight(border)
				},
			),
			view.FlexGrow("1"),
		).FlexGrow("1"),
	).Height(browser.Size{Value: 100, Unit: browser.UnitVH})
}

func header(s *State) *browser.Node {
	return ui.HStack(
		ui.HStack(
			icons.Trademark(s.Theme.BackgroundColor, 30),
			ui.OnlyIf(s.SidebarHidden,
				func() *browser.Node { return s.Theme.Text(s.SidebarState.SelectedDisplay) },
			),
		).OnClickDispatch(EventToggleSidebar{}).
			FlexGrow("1").AlignItemsCenter(),
		s.Theme.Textf("v%s", s.ClientVersion).
			FontSize(browser.Size{Value: 10, Unit: browser.UnitPX}).
			MarginRight(browser.Size{Value: 15, Unit: browser.UnitPX}),
	).BorderBottom(browser.Border{
		Width: browser.Size{Value: 1, Unit: browser.UnitPX},
		Type:  browser.BorderSolid,
		Color: "lightgray",
	},
	).AlignItemsCenter().CursorPointer()
}

var border = browser.Border{
	Width: browser.Size{Value: 1, Unit: browser.UnitPX},
	Type:  browser.BorderSolid,
	Color: "lightgray",
}

var (
	DarkTheme = ui.Theme{
		//	FontFamily:           "monospace",
		BackgroundColor:      "black",
		HoverBackgroundColor: "rgb(30,30,30)",
		TextColor:            "antiquewhite",
		//		LinkColor:       "rgb(33,95,180)",
		LinkColor: "rgb(66,196,208)",
	}
	LightTheme = ui.Theme{
		//	FontFamily:           "monospace",
		BackgroundColor:      "white",
		HoverBackgroundColor: "rgb(238,238,238)",
		TextColor:            "black",
		LinkColor:            "blue",
	}
	DefaultTheme = LightTheme
)
