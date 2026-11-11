package main

import (
	"fmt"
	"testing"

	"aibo-rev/ui-checks/acceptance"
)

// prototypeTarget 让原型接受 tools/ui-checks 的验收套件。原型只提供 idle-empty 与 history-10k。
type prototypeTarget struct{}

func (prototypeTarget) Open(fixture string) (acceptance.Fixture, error) {
	switch fixture {
	case "idle-empty":
		app := newApp(0)
		app.items = nil
		return acceptance.Fixture{View: app.view}, nil
	case "history-10k":
		app := newApp(10000)
		app.loadAll()
		var reply *Item
		runes := []rune(longReply)
		i := 0
		stream := func() bool {
			if reply == nil {
				reply = app.append(&Item{Kind: KindAgent})
				app.list.ScrollToEnd()
			}
			if i >= len(runes) {
				reply.Done = true
				return false
			}
			i = min(len(runes), i+6)
			reply.Text = string(runes[:i])
			return true
		}
		return acceptance.Fixture{View: app.view, Stream: stream}, nil
	}
	for _, f := range acceptance.Fixtures {
		if f == fixture {
			return acceptance.Fixture{}, acceptance.ErrUnsupported
		}
	}
	return acceptance.Fixture{}, fmt.Errorf("未知夹具 %q", fixture)
}

func TestAcceptance(t *testing.T) { acceptance.Run(t, prototypeTarget{}) }
