package render

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
)

// Binder interface for managing request payloads.
type Binder interface {
	Bind(*http.Request) error
}

// Decoder interface for custom request decoders.
type Decoder interface {
	Decode(*http.Request) error
}

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "render context value " + k.name
}

var (
	// StrictDecoderKey is the context key to enable strict decoding.
	StrictDecoderKey = &contextKey{"StrictDecoder"}
)

// StrictDecoder is a middleware that enables strict JSON decoding.
func StrictDecoder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), StrictDecoderKey, true))
		next.ServeHTTP(w, r)
	})
}

// Decode binds the request body to the destination struct.
func Decode(r *http.Request, v interface{}) error {
	var err error

	// Check if the destination struct implements Decoder.
	if decoder, ok := v.(Decoder); ok {
		err = decoder.Decode(r)
	} else {
		err = DefaultDecoder(r, v)
	}

	if err != nil {
		return err
	}

	// Check if the destination struct implements Binder.
	if binder, ok := v.(Binder); ok {
		err = binder.Bind(r)
	}

	return err
}

// DefaultDecoder detects the Content-Type and decodes the payload.
func DefaultDecoder(r *http.Request, v interface{}) error {
	var err error

	switch GetRequestContentType(r) {
	case ContentJSON:
		err = DecodeJSON(r, v)
	case ContentXML:
		err = DecodeXML(r, v)
	case ContentForm:
		err = DecodeForm(r, v)
	case ContentMultipartForm:
		// TODO: DecodeMultipartForm(r, v)
	default:
		err = errors.New("render: unable to automatically decode the request content type")
	}

	return err
}

// DecodeJSON decodes a JSON request body into the destination struct.
func DecodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	if isStrict, ok := r.Context().Value(StrictDecoderKey).(bool); ok && isStrict {
		dec.DisallowUnknownFields()
	}
	return dec.Decode(v)
}

// DecodeXML decodes an XML request body into the destination struct.
func DecodeXML(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return xml.NewDecoder(r.Body).Decode(v)
}

// DecodeForm decodes a form request body into the destination struct.
func DecodeForm(r *http.Request, v interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	if err := DecodePostForm(r, v); err != nil {
		return err
	}
	return DecodeQuery(r, v)
}

// DecodePostForm decodes a post form request body into the destination struct.
func DecodePostForm(r *http.Request, v interface{}) error {
	decoder := NewFormDecoder(r.PostForm)
	return decoder.Decode(v)
}

// DecodeQuery decodes a query string into the destination struct.
func DecodeQuery(r *http.Request, v interface{}) error {
	decoder := NewFormDecoder(r.URL.Query())
	return decoder.Decode(v)
}