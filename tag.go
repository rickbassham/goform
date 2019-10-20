package goform

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type flags struct {
	base64   bool
	required bool
}

func parseTag(tag string) (string, flags) {
	split := strings.Split(tag, ",")

	if len(split) > 0 {
		var f flags

		for _, option := range split[1:] {
			switch option {
			case "base64":
				f.base64 = true
			case "required":
				f.required = true
			}
		}

		return split[0], f
	}

	return "", flags{}
}

func base(tag reflect.StructTag) (int, error) {
	base := int64(10)

	if baseStr, ok := tag.Lookup("base"); ok {
		var err error
		base, err = strconv.ParseInt(baseStr, 10, 32)
		if err != nil {
			return int(base), errors.New("invalid int base")
		}
	}

	return int(base), nil
}
