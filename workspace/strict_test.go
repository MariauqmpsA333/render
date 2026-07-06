package render

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type StrictTestStruct struct {
	KnownField string `json:"known_field"`
}

type NestedStruct struct {
	Inner KnownInner `json:"inner"`
}

type KnownInner struct {
	Value string `json:"value"`
}

func TestStrictDecoding_Failure(t *testing.T) {
	h := StrictDecoder(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data StrictTestStruct
		err := Decode(r, &data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	ts := httptest.NewServer(h)
	defer ts.Close()

	payload := `{"known_field": "value", "unknown_field": "value"}`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "unknown field") {
		t.Errorf("expected error containing 'unknown field', got %q", string(body))
	}
}

func TestStrictDecoding_Success(t *testing.T) {
	h := StrictDecoder(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data StrictTestStruct
		err := Decode(r, &data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	ts := httptest.NewServer(h)
	defer ts.Close()

	payload := `{"known_field": "value"}`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestNonStrictDecoding_Success(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data StrictTestStruct
		err := Decode(r, &data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(h)
	defer ts.Close()

	payload := `{"known_field": "value", "unknown_field": "value"}`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestStrictDecoding_NestedFailure(t *testing.T) {
	h := StrictDecoder(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data NestedStruct
		err := Decode(r, &data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	ts := httptest.NewServer(h)
	defer ts.Close()

	payload := `{"inner": {"value": "val", "unknown_inner_field": "val"}}`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "unknown field") {
		t.Errorf("expected error containing 'unknown field', got %q", string(body))
	}
}