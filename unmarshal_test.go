package goform_test

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rickbassham/goform"
)

func TestUnmarshal_URLEncoded(t *testing.T) {
	data := url.Values{}
	data.Set("name", "rick")
	data.Set("age", "39")

	r, err := http.NewRequest(http.MethodPost, "http://test/page?id=1", strings.NewReader(data.Encode()))
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	type body struct {
		ID   int    `form:"id"`
		Name string `form:"name"`
		Age  int    `form:"age"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	require.NoError(t, err)

	assert.Equal(t, body{
		ID:   1,
		Name: "rick",
		Age:  39,
	}, b)
}

func TestUnmarshal_MultiPartForm(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	writeFormField(w, "id", "1")
	writeFormField(w, "name", "rick")
	writeFormField(w, "age", "39")

	w.Close() // nolint

	r, err := http.NewRequest(http.MethodPost, "http://test/page", &buf)
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", w.FormDataContentType())

	type body struct {
		ID   int    `form:"id"`
		Name string `form:"name"`
		Age  int    `form:"age"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	require.NoError(t, err)

	assert.Equal(t, body{
		ID:   1,
		Name: "rick",
		Age:  39,
	}, b)
}

func TestUnmarshal_MultiPartFormImage(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	headshot := image.NewGray16(image.Rect(0, 0, 32, 32))
	draw.Draw(headshot, image.Rect(8, 8, 24, 24), image.NewUniform(color.Gray16{128}), image.Point{0, 0}, draw.Over)

	var imgBuf bytes.Buffer
	png.Encode(&imgBuf, headshot) // nolint

	writeFormField(w, "id", "1")
	writeFormField(w, "name", "rick")
	writeFormField(w, "age", "39")
	writeFormFile(w, "headshot", &imgBuf)

	w.Close() // nolint

	r, err := http.NewRequest(http.MethodPost, "http://test/page", &buf)
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", w.FormDataContentType())

	type body struct {
		ID       int         `form:"id"`
		Name     string      `form:"name"`
		Headshot image.Image `form:"headshot"`
		Age      int         `form:"age"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	require.NoError(t, err)

	assert.Equal(t, body{
		ID:       1,
		Name:     "rick",
		Age:      39,
		Headshot: headshot,
	}, b)
}

func writeFormField(w *multipart.Writer, name, value string) {
	vw, _ := w.CreateFormField(name)
	vw.Write([]byte(value)) // nolint
}

func writeFormFile(w *multipart.Writer, name string, value io.Reader) {
	vw, _ := w.CreateFormFile(name, name)
	io.Copy(vw, value) // nolint
}

func TestUnmarshal_MultiPartFormImageBase64(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	headshot := image.NewGray16(image.Rect(0, 0, 32, 32))
	draw.Draw(headshot, image.Rect(8, 8, 24, 24), image.NewUniform(color.Gray16{128}), image.Point{0, 0}, draw.Over)

	var imgBuf bytes.Buffer
	bw := base64.NewEncoder(base64.StdEncoding, &imgBuf)
	png.Encode(bw, headshot) // nolint
	bw.Close()               // nolint

	writeFormField(w, "id", "1")
	writeFormField(w, "name", "rick")
	writeFormField(w, "age", "39")
	writeFormFile(w, "headshot", &imgBuf)

	w.Close() // nolint

	r, err := http.NewRequest(http.MethodPost, "http://test/page", &buf)
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", w.FormDataContentType())

	type body struct {
		ID       int         `form:"id"`
		Name     string      `form:"name"`
		Age      int         `form:"age"`
		Headshot image.Image `form:"headshot,base64"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	require.NoError(t, err)

	assert.Equal(t, body{
		ID:       1,
		Name:     "rick",
		Age:      39,
		Headshot: headshot,
	}, b)
}

func TestUnmarshal_MultiPartFormByteSlice(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	writeFormField(w, "id", "1")
	writeFormField(w, "name", "rick")
	writeFormField(w, "age", "39")
	writeFormField(w, "headshot", "ABCD")

	w.Close() // nolint

	r, err := http.NewRequest(http.MethodPost, "http://test/page", &buf)
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", w.FormDataContentType())

	type body struct {
		ID       int    `form:"id"`
		Name     string `form:"name"`
		Age      int    `form:"age"`
		Headshot []byte `form:"headshot"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	require.NoError(t, err)

	assert.Equal(t, body{
		ID:       1,
		Name:     "rick",
		Age:      39,
		Headshot: []byte("ABCD"),
	}, b)
}

func TestUnmarshal_MultiPartFormImageByteSlice(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	headshot := image.NewGray16(image.Rect(0, 0, 32, 32))
	draw.Draw(headshot, image.Rect(8, 8, 24, 24), image.NewUniform(color.Gray16{128}), image.Point{0, 0}, draw.Over)

	var imgBuf bytes.Buffer
	bw := base64.NewEncoder(base64.StdEncoding, &imgBuf)
	png.Encode(bw, headshot) // nolint
	bw.Close()               // nolint

	var copyBuf bytes.Buffer
	png.Encode(&copyBuf, headshot) // nolint

	writeFormField(w, "id", "1")
	writeFormField(w, "name", "rick")
	writeFormField(w, "age", "39")
	writeFormFile(w, "headshot", &imgBuf)

	w.Close() // nolint

	r, err := http.NewRequest(http.MethodPost, "http://test/page", &buf)
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", w.FormDataContentType())

	type body struct {
		ID       int    `form:"id"`
		Name     string `form:"name"`
		Age      int    `form:"age"`
		Headshot []byte `form:"headshot,base64"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	require.NoError(t, err)

	assert.Equal(t, body{
		ID:       1,
		Name:     "rick",
		Age:      39,
		Headshot: copyBuf.Bytes(),
	}, b)
}

func TestUnmarshal_JSON(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://test/page", strings.NewReader(`{"id": 1, "name": "rick", "age": 39}`))
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", "application/json; charset=utf-8")

	type body struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	require.NoError(t, err)

	assert.Equal(t, body{
		ID:   1,
		Name: "rick",
		Age:  39,
	}, b)
}

func TestUnmarshal_QueryStringAndJSON(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://test/page?something=abc", strings.NewReader(`{"id": 1, "name": "rick", "age": 39}`))
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", "application/json")

	type body struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Age       int    `json:"age"`
		Something string `form:"something"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	require.NoError(t, err)

	assert.Equal(t, body{
		ID:        1,
		Name:      "rick",
		Age:       39,
		Something: "abc",
	}, b)
}

func TestUnmarshal_QueryStringAndJSONOverride(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://test/page?something=abc", strings.NewReader(`{"id": 1, "name": "rick", "age": 39}`))
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", "application/json")

	type body struct {
		ID   int    `json:"id"`
		Name string `json:"name" form:"something"`
		Age  int    `json:"age"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	require.NoError(t, err)

	assert.Equal(t, body{
		ID:   1,
		Name: "abc",
		Age:  39,
	}, b)
}

func TestUnmarshal_RequiredMissing(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://test/page", strings.NewReader(`{"id": 1, "name": "rick", "age": 39}`))
	require.NoError(t, err)
	require.NotNil(t, r)

	r.Header.Add("Content-Type", "application/json")

	type body struct {
		Something string `form:"something,required"`
	}

	var b body

	err = goform.Unmarshal(r, &b)
	assert.EqualError(t, err, "goform: missing required field [something]")
}
