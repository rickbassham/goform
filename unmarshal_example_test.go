package goform_test

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/rickbassham/goform"
)

func ExampleUnmarshal() {
	data := url.Values{}
	data.Set("name", "rick")
	data.Set("age", "39")

	r, err := http.NewRequest(http.MethodPost, "http://test/page?id=1", strings.NewReader(data.Encode()))
	if err != nil {
		panic(err.Error())
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	type body struct {
		ID   int    `form:"id"`
		Name string `form:"name"`
		Age  int    `form:"age"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	if err != nil {
		panic(err.Error())
	}
}
