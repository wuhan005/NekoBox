// Copyright 2025 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"encoding/json"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"

	"github.com/flamego/flamego"
	"github.com/unknwon/com"
	"github.com/wuhan005/govalid"
	"golang.org/x/text/language"

	"github.com/wuhan005/NekoBox/internal/context"
)

func Bind(model interface{}) flamego.Handler {
	// Ensure not pointer.
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		panic("form: pointer can not be accepted as binding model")
	}

	return func(ctx context.Context) error {
		obj := reflect.New(reflect.TypeOf(model))
		r := ctx.Request().Request
		if r.Body != nil {
			defer func() { _ = r.Body.Close() }()
			if err := json.NewDecoder(r.Body).Decode(obj.Interface()); err != nil {
				return ctx.Error(http.StatusBadRequest, "表单解析失败")
			}
		}

		acceptLanguage := r.Header.Get("Accept-Language")
		languageTags, _, _ := language.ParseAcceptLanguage(acceptLanguage)
		languageTag := language.Chinese
		if len(languageTags) > 0 {
			languageTag = languageTags[0]
		}

		errors, ok := govalid.Check(obj.Interface(), languageTag)
		if !ok {
			var msg string
			if len(errors) > 0 {
				msg = errors[0].Error()
			}
			return ctx.Error(http.StatusBadRequest, "%s", msg)
		}

		// Validation passed.
		ctx.Map(obj.Elem().Interface())
		return nil
	}
}

func BindMultipart(model interface{}) flamego.Handler {
	// Ensure not pointer.
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		panic("form: pointer can not be accepted as binding model")
	}

	return func(ctx context.Context) error {
		obj := reflect.New(reflect.TypeOf(model))

		r := ctx.Request()

		contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		switch contentType {
		case "application/x-www-form-urlencoded", "":
			if err := r.ParseForm(); err != nil {
				return ctx.Error(http.StatusBadRequest, "表单解析失败")
			}
		case "multipart/form-data":
			if err := r.ParseMultipartForm(10 * 1 << 20); err != nil { // 10 MiB
				return ctx.Error(http.StatusBadRequest, "表单解析失败")
			}
		default:
			return ctx.Error(http.StatusBadRequest, "不支持的 Content-Type")
		}

		// Bind the form data to the given struct.
		typ := reflect.TypeOf(obj.Interface())
		val := reflect.ValueOf(obj.Interface())
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			val = val.Elem()
		}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldName := typ.Field(i).Tag.Get("form")
			if fieldName == "" {
				fieldName = com.ToSnakeCase(field.Name)
			}

			if r.MultipartForm != nil {
				fhType := reflect.TypeOf((*multipart.FileHeader)(nil))

				value, ok := r.MultipartForm.Value[fieldName]
				if ok {
					fieldValue := reflect.ValueOf(value[0])
					switch val.Field(i).Kind() {
					case reflect.Bool:
						val.Field(i).SetBool(fieldValue.String() == "true" || fieldValue.String() == "on" || fieldValue.String() == "1")
						continue
					case reflect.Slice:
						numElems := len(value)
						slice := reflect.MakeSlice(val.Field(i).Type(), numElems, numElems)
						for j := 0; j < numElems; j++ {
							slice.Index(j).Set(reflect.ValueOf(value[j]))
						}
						val.Field(i).Set(slice)
						continue
					default:
						val.Field(i).Set(fieldValue)
					}

				} else {
					value, ok := r.MultipartForm.File[fieldName]
					if ok {
						numElems := len(value)

						if val.Field(i).Kind() == reflect.Slice && numElems > 0 && val.Field(i).Type().Elem() == fhType {
							slice := reflect.MakeSlice(val.Field(i).Type(), numElems, numElems)
							for i := 0; i < numElems; i++ {
								slice.Index(i).Set(reflect.ValueOf(value[i]))
							}
							val.Field(i).Set(slice)
						} else if val.Field(i).Type() == fhType {
							val.Field(i).Set(reflect.ValueOf(value[0]))
						}
					}
				}
			} else {
				val.Field(i).Set(reflect.ValueOf(r.Form.Get(fieldName)))
			}
		}

		acceptLanguage := r.Header.Get("Accept-Language")
		languageTags, _, _ := language.ParseAcceptLanguage(acceptLanguage)
		languageTag := language.Chinese
		if len(languageTags) > 0 {
			languageTag = languageTags[0]
		}

		errors, ok := govalid.Check(obj.Interface(), languageTag)
		if !ok {
			var msg string
			if len(errors) > 0 {
				msg = errors[0].Error()
			}
			return ctx.Error(http.StatusBadRequest, "%s", msg)
		}

		// Validation passed.
		ctx.Map(obj.Elem().Interface())
		return nil
	}
}
