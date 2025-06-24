package bindings

import (
	"cmp"
	"slices"

	"github.com/ufukty/gohandlers/pkg/inspects"
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
		return cmp.Or(cmp.Compare(a.receiver.Type, b.receiver.Type), cmp.Compare(a.handler, b.handler))
	})
	return o
}
