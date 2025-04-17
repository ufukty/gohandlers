package basics

import (
	"fmt"
	"strconv"
)

type Boolean string

const (
	True  Boolean = "true"
	False Boolean = "false"
)

func (b *Boolean) FromRoute(src string) error {
	*b = Boolean(src)
	return nil
}

func (b Boolean) ToRoute() (string, error) {
	return string(b), nil
}

func (s *Boolean) FromQuery(v string) error {
	*s = Boolean(v)
	return nil
}

func (s Boolean) ToQuery() (string, bool, error) {
	return string(s), s != "", nil
}

func (b Boolean) Validate() error {
	switch b {
	case False:
		return nil
	case True:
		return nil
	}
	return fmt.Errorf("invalid value: %q", b)
}

type String string

func (s *String) FromRoute(v string) error {
	*s = String(v)
	return nil
}

func (s String) ToRoute() (string, error) {
	return string(s), nil
}

func (s *String) FromQuery(v string) error {
	*s = String(v)
	return nil
}

func (s String) ToQuery() (string, bool, error) {
	return string(s), s != "", nil
}

type Int int

func (i *Int) FromRoute(v string) error {
	integer, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("strconv.Atoi: %w", err)
	}
	*i = Int(integer)
	return nil
}

func (i Int) ToRoute() (string, error) {
	return strconv.Itoa(int(i)), nil
}

func (i *Int) FromQuery(v string) error {
	integer, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("strconv.Atoi: %w", err)
	}
	*i = Int(integer)
	return nil
}

func (i Int) ToQuery() (string, bool, error) {
	return strconv.Itoa(int(i)), i != 0, nil
}
