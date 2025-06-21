package basics

import (
	"fmt"
	"strconv"
)

type Boolean bool

const (
	True  Boolean = true
	False Boolean = false
)

func (b *Boolean) FromRoute(v string) error {
	*b = v == "t"
	return nil
}

func (b Boolean) ToRoute() (string, error) {
	if b {
		return "t", nil
	}
	return "f", nil
}

func (b *Boolean) FromQuery(v string) error {
	*b = v == "t"
	return nil
}

func (b Boolean) ToQuery() (string, bool, error) {
	if b {
		return "t", true, nil
	}
	return "f", true, nil
}

func (b Boolean) Validate() any { return nil }

type FormBoolean bool

const (
	On  FormBoolean = true
	Off FormBoolean = false
)

func (b *FormBoolean) FromRoute(v string) error {
	*b = v == "on"
	return nil
}

func (b FormBoolean) ToRoute() (string, error) {
	if b {
		return "on", nil
	}
	return "f", nil
}

func (b *FormBoolean) FromQuery(v string) error {
	*b = v == "on"
	return nil
}

func (b FormBoolean) ToQuery() (string, bool, error) {
	if b {
		return "on", true, nil
	}
	return "", false, nil
}

func (b FormBoolean) Validate() any { return nil }

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

func (s String) Validate() any { return nil }

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

func (i Int) Validate() any { return nil }
