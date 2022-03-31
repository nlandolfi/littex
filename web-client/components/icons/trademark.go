package icons

import (
	"github.com/spinsrv/browser"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// go:embed trademark-black.svg
// var trademarkBlockSVG string
// maybe embed in future

var trademarkStyle = browser.Style{
	Width:   browser.Size{Value: 35, Unit: browser.UnitPX},
	Height:  browser.Size{Value: 35, Unit: browser.UnitPX},
	Padding: browser.Size{Value: 5, Unit: browser.UnitPX},
}

func TrademarkBlack(si float64) *browser.Node {
	s := trademarkStyle
	s.Width = browser.Size{Value: si, Unit: browser.UnitPX}
	s.Height = browser.Size{Value: si, Unit: browser.UnitPX}
	return &browser.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Img,
		Style:    s,
		Attr: []*html.Attribute{
			&html.Attribute{
				Key: "src",
				Val: "trademark-black.svg",
			},
		},
	}
}

func TrademarkWhite(si float64) *browser.Node {
	s := trademarkStyle
	s.Width = browser.Size{Value: si, Unit: browser.UnitPX}
	s.Height = browser.Size{Value: si, Unit: browser.UnitPX}

	return &browser.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Img,
		Style:    s,
		Attr: []*html.Attribute{
			&html.Attribute{
				Key: "src",
				Val: "trademark-white.svg",
			},
		},
	}
}

func Trademark(backgroundColor string, si float64) *browser.Node {
	var tm *browser.Node
	switch backgroundColor {
	case "white":
		tm = TrademarkBlack(si)
	case "black":
		tm = TrademarkWhite(si)
	default:
		tm = TrademarkBlack(si)
	}
	return tm
}
