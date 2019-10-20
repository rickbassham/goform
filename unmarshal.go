// Package goform is meant to make binding http data to structs easy.
package goform

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

var (
	defaultMaxMemory int64 = 32 << 20 // 32 MB
)

// Unmarshal will bind the body and query string values to the given struct.
// Works will all primitive types, time.Time, image.Image, and []byte.
// It first inspects the Content-Type header of the request. If the Content-Type
// is json it will use the json.Unmarshal func and then bind anything from the
// query string as well.
func Unmarshal(r *http.Request, v interface{}) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	defer r.Body.Close()

	if mediaType == "application/json" {
		err = json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			return err
		}
	}

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return errors.New("goform: v must be a pointer")
	}

	t = t.Elem()
	val := reflect.Indirect(reflect.ValueOf(v))

	r.ParseMultipartForm(defaultMaxMemory) // nolint

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag, tagOptions := parseTag(f.Tag.Get("form"))

		if tag == "" || tag == "-" {
			continue
		}

		valf := val.FieldByName(f.Name)
		kind := f.Type.Kind()

		if kind == reflect.Ptr {
			kind = f.Type.Elem().Kind()
			valf.Set(reflect.New(f.Type.Elem()))
			valf = reflect.Indirect(valf)
		}

		formValues := r.Form[tag]

		if len(formValues) > 1 {
			return errors.New("goform: arrays not supported yet")
		}

		if len(formValues) == 0 {
			err = decodeMultipart(r, tag, valf, kind, tagOptions)
			if err != nil {
				return err
			}

			// formValues is empty, so just move along
			continue
		}

		formValue := formValues[0]

		err = decodeFormValue(valf, kind, f, formValue)
		if err != nil {
			return err
		}

	}

	return nil
}

func decodeFormValue(valf reflect.Value, kind reflect.Kind, f reflect.StructField, formValue string) error {
	var err error

	switch kind {
	case reflect.Slice:
		if valf.Type() == reflect.TypeOf([]byte{}) {
			valf.SetBytes([]byte(formValue))
		}
		break
	case reflect.String:
		valf.SetString(formValue)
	case reflect.Bool:
		err = decodeBool(valf, formValue)
	case reflect.Int:
		err = decodeInt(valf, f.Tag, 0, formValue)
	case reflect.Int8:
		err = decodeInt(valf, f.Tag, 8, formValue)
	case reflect.Int16:
		err = decodeInt(valf, f.Tag, 16, formValue)
	case reflect.Int32:
		err = decodeInt(valf, f.Tag, 32, formValue)
	case reflect.Int64:
		err = decodeInt(valf, f.Tag, 64, formValue)
	case reflect.Uint:
		err = decodeUint(valf, f.Tag, 0, formValue)
	case reflect.Uint8:
		err = decodeUint(valf, f.Tag, 8, formValue)
	case reflect.Uint16:
		err = decodeUint(valf, f.Tag, 16, formValue)
	case reflect.Uint32:
		err = decodeUint(valf, f.Tag, 32, formValue)
	case reflect.Uint64:
		err = decodeUint(valf, f.Tag, 64, formValue)
	case reflect.Float32:
		err = decodeFloat(valf, 32, formValue)
	case reflect.Float64:
		err = decodeFloat(valf, 64, formValue)
	case reflect.Struct:
		err = decodeStruct(valf, f, formValue)
	default:
		err = errors.New("goform: invalid destination type")
	}

	return err
}

func decodeBool(valf reflect.Value, value string) error {
	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	valf.SetBool(boolVal)

	return nil
}

func decodeFloat(valf reflect.Value, bitSize int, value string) error {
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return err
	}
	valf.SetFloat(floatVal)

	return nil
}

func decodeInt(valf reflect.Value, tag reflect.StructTag, bitSize int, value string) error {
	b, err := base(tag)
	if err != nil {
		return err
	}

	intVal, err := strconv.ParseInt(value, b, bitSize)
	if err != nil {
		return err
	}
	valf.SetInt(intVal)

	return nil
}

func decodeUint(valf reflect.Value, tag reflect.StructTag, bitSize int, value string) error {
	b, err := base(tag)
	if err != nil {
		return err
	}

	intVal, err := strconv.ParseUint(value, b, bitSize)
	if err != nil {
		return err
	}
	valf.SetUint(intVal)

	return nil
}

func decodeStruct(valf reflect.Value, f reflect.StructField, formValue string) error {
	if valf.Type() == reflect.TypeOf(time.Time{}) {
		format := f.Tag.Get("format")
		if format == "" {
			format = time.RFC3339
		}

		var timeVal time.Time
		var err error

		tz := f.Tag.Get("tz")
		if tz == "" {
			timeVal, err = time.Parse(format, formValue)
		} else {
			var loc *time.Location
			loc, err = time.LoadLocation(tz)
			if err != nil {
				return err
			}

			timeVal, err = time.ParseInLocation(format, formValue, loc)
		}
		if err != nil {
			return err
		}
		valf.Set(reflect.ValueOf(timeVal))
	} else {
		return errors.New("goform: invalid destination type")
	}

	return nil
}

func decodeMultipart(r *http.Request, tag string, valf reflect.Value, kind reflect.Kind, tagOptions flags) error {
	if r.MultipartForm != nil {
		headers := r.MultipartForm.File[tag]
		if len(headers) == 0 {
			if tagOptions.required {
				return fmt.Errorf("goform: missing required field [%s]", tag)
			}

			return nil
		}

		err := decodeMultipartFile(valf, kind, tagOptions, headers[0])
		if err != nil {
			return err
		}

		return nil
	}

	if tagOptions.required {
		return fmt.Errorf("goform: missing required field [%s]", tag)
	}

	return nil
}

func decodeMultipartFile(valf reflect.Value, kind reflect.Kind, tagOptions flags, hdr *multipart.FileHeader) error {
	var rdr io.Reader
	var err error

	data, err := hdr.Open()
	if err != nil {
		return err
	}
	defer data.Close()

	rdr = data

	if tagOptions.base64 {
		rdr = base64.NewDecoder(base64.StdEncoding, rdr)
	}

	if valf.Type() == reflect.TypeOf([]byte{}) {
		readData, err := ioutil.ReadAll(rdr)
		if err != nil {
			return err
		}

		valf.SetBytes(readData)
		return nil
	} else if valf.Type().Implements(reflect.TypeOf((*image.Image)(nil)).Elem()) {
		var img image.Image

		img, _, err = image.Decode(rdr)
		if err != nil {
			return err
		}

		valf.Set(reflect.ValueOf(img))
		return nil
	}

	return nil
}
