package bindings

import (
	"gohandlers/pkg/inspects"
	"slices"
)

// just used for stable ordering of methods
type funcrecv struct {
	receiver inspects.Receiver // receiver of the handler. not the receiver of Parse and Build methods
	handler  string
}

func ordered(infoss map[inspects.Receiver]map[string]inspects.Info) []funcrecv {
	o := []funcrecv{}
	for recv, handlers := range infoss {
		for handler := range handlers {
			o = append(o, funcrecv{recv, handler})
		}
	}
	slices.SortFunc(o, func(a, b funcrecv) int {
		if a.receiver.Type < b.receiver.Type {
			return -1
		}
		if a.receiver.Type > b.receiver.Type {
			return 1
		}

		if a.handler < b.handler {
			return -1
		}
		if a.handler > b.handler {
			return 1
		}

		return 0
	})
	return o
}
