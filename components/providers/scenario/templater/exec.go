package templater

import (
	"fmt"

	"github.com/yandex/pandora/lib/mp"
)

var ErrUnsupportedFunctionType = fmt.Errorf("unsupported function type")

func ExecTemplateFuncWithVariables(fun any, args []string, templateVars map[string]any, iter mp.Iterator) (string, error) {
	a := make([]any, len(args))
	for i := range args {
		v, err := mp.GetMapValue(templateVars, args[i], iter)
		if err == nil {
			a[i] = v
		} else {
			a[i] = args[i]
		}
	}
	switch exec := fun.(type) {
	case func() string:
		ans := exec()
		return ans, nil
	case func(args ...any) string:
		ans := exec(a...)
		return ans, nil
	}
	return "", ErrUnsupportedFunctionType
}

func ExecTemplateFunc(fun any, args []string, templateVars map[string]any, iter mp.Iterator) (string, error) {
	a := make([]any, len(args))
	for i := range args {
		a[i] = args[i]
	}
	switch exec := fun.(type) {
	case func() string:
		ans := exec()
		return ans, nil
	case func(args ...any) string:
		ans := exec(a...)
		return ans, nil
	}
	return "", ErrUnsupportedFunctionType
}
