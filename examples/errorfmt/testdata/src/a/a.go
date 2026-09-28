package a

import (
	"errors"
	"fmt"
)

func f(name string) error {
	return errors.New(fmt.Sprintf("not found: %s", name)) // want `use fmt\.Errorf`
}

func g(name string) error {
	msg := fmt.Sprintf("not found: %s", name)
	return errors.New(msg) // 一度変数に入れたものは対象外
}

func h(name string) error {
	return fmt.Errorf("not found: %s", name)
}
