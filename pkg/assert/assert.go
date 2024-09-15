package assert

import (
	"encoding/json"
	"fmt"
)

func stringify(item any) string {
	if item == nil {
		return "nil"
	}

	switch t := item.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case int:
		return fmt.Sprintf("%d", item)
	default:
		d, err := json.Marshal(t)
		if err != nil {
			return string(d)
		}
	}
	return fmt.Sprintf("%s", item)
}

func Assert(truth bool, msg string, data ...any) {
	if !truth {
		for _, item := range data {
			fmt.Println(stringify(item))
		}

		panic(msg)
	}
}
