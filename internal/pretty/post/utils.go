package post

import "go/token"

type actions struct {
	newline, comma bool
}

func (a actions) merge(b actions) actions {
	return actions{
		newline: a.newline || b.newline,
		comma:   a.comma || b.comma,
	}
}

func concat(ms ...map[token.Pos]actions) map[token.Pos]actions {
	c := map[token.Pos]actions{}
	for _, m := range ms {
		if m == nil {
			continue
		}
		for k, v := range m {
			if _, ok := c[k]; !ok {
				c[k] = v
			} else {
				c[k] = c[k].merge(v)
			}
		}
	}
	return c
}
