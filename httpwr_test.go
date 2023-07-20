package httpwr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeNoError(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	F(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}).ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected http status ok, got %d", resp.StatusCode)
	}
}

func TestServeHTTPError(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	status := http.StatusBadRequest
	msg := "data was wrong"
	F(func(w http.ResponseWriter, r *http.Request) error {
		return Error{
			Status: status,
			Err:    fmt.Errorf(msg),
		}
	}).ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != status {
		t.Fatalf("expected http status %d, got %d", status, resp.StatusCode)
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if !strings.Contains(string(bts), msg) {
		t.Fatalf("%q does not contain %q", string(bts), msg)
	}
}

func TestServeError(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	msg := "server is doing funky stuff"
	F(func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf(msg)
	}).ServeHTTP(w, req)
	resp := w.Result()

	status := http.StatusInternalServerError
	if resp.StatusCode != status {
		t.Fatalf("expected http status %d, got %d", status, resp.StatusCode)
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if !strings.Contains(string(bts), msg) {
		t.Fatalf("%q does not contain %q", string(bts), msg)
	}
}

func TestWrap(t *testing.T) {
	type args struct {
		err    error
		status int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil err",
			args: args{
				err:    nil,
				status: 400,
			},
			wantErr: false,
		},
		{
			name: "nonnil err",
			args: args{
				err:    fmt.Errorf("errr"),
				status: 404,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.args.status, tt.args.err)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	t.Run("outer", func(t *testing.T) {
		if !errors.Is(Wrap(http.StatusGone, io.EOF), Error{}) {
			t.Fatalf("underlying error should be Error")
		}
	})

	t.Run("is", func(t *testing.T) {
		if !errors.Is(Wrap(http.StatusGone, io.EOF), io.EOF) {
			t.Fatalf("underlying error should be io.EOF")
		}
	})

	t.Run("is not", func(t *testing.T) {
		if errors.Is(Wrap(http.StatusGone, io.ErrUnexpectedEOF), io.EOF) {
			t.Fatalf("underlying error should not be io.EOF")
		}
	})
}

func TestErrorf(t *testing.T) {
	err := Errorf(http.StatusConflict, "foo bar %d", 10)
	expectedErr := Error{
		Err:    fmt.Errorf("foo bar 10"),
		Status: http.StatusConflict,
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("errors does not match: %v", err)
	}
}

func TestOK(t *testing.T) {
	msg := "some msg"

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	F(func(w http.ResponseWriter, r *http.Request) error {
		return OK(w, http.StatusOK, msg)
	}).ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected http status ok, got %d", resp.StatusCode)
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if !strings.Contains(string(bts), msg) {
		t.Fatalf("%q does not contain %q", string(bts), msg)
	}
}

func TestOKWithData(t *testing.T) {
	data := M{
		"some": "data",
		"age":  23,
	}

	msg := "some msg"

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	F(func(w http.ResponseWriter, r *http.Request) error {
		return OKWithData(w, http.StatusOK, msg, data)
	}).ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected http status ok, got %d", resp.StatusCode)
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if !strings.Contains(string(bts), msg) {
		t.Fatalf("%q does not contain %q", string(bts), msg)
	}

	mapJson, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("map to json error: %v", err)
	}
	if !strings.Contains(string(bts), string(mapJson)) {
		t.Fatalf("%q does not contain %q", string(bts), string(mapJson))
	}
}

func TestConstMessage(t *testing.T) {
	t.Parallel()

	t.Run("created", func(t *testing.T) {
		msg := CreatedMsg

		req := httptest.NewRequest("GET", "/created", nil)
		w := httptest.NewRecorder()
		F(func(w http.ResponseWriter, r *http.Request) error {
			return OK(w, http.StatusCreated, msg)
		}).ServeHTTP(w, req)
		resp := w.Result()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected http status ok, got %d", resp.StatusCode)
		}

		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}
		if !strings.Contains(string(bts), msg) {
			t.Fatalf("%q does not contain %q", string(bts), msg)
		}
	})

	t.Run("ok", func(t *testing.T) {
		msg := OKMsg

		req := httptest.NewRequest("GET", "/ok", nil)
		w := httptest.NewRecorder()
		F(func(w http.ResponseWriter, r *http.Request) error {
			return OK(w, http.StatusOK, msg)
		}).ServeHTTP(w, req)
		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected http status ok, got %d", resp.StatusCode)
		}

		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}
		if !strings.Contains(string(bts), msg) {
			t.Fatalf("%q does not contain %q", string(bts), msg)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		msg := InternalServerErrorMsg

		req := httptest.NewRequest("GET", "/internalerr", nil)
		w := httptest.NewRecorder()
		F(func(w http.ResponseWriter, r *http.Request) error {
			return OK(w, http.StatusOK, msg)
		}).ServeHTTP(w, req)
		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected http status ok, got %d", resp.StatusCode)
		}

		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}
		if !strings.Contains(string(bts), msg) {
			t.Fatalf("%q does not contain %q", string(bts), msg)
		}
	})

	t.Run("bad request", func(t *testing.T) {
		msg := BadRequestMsg

		req := httptest.NewRequest("GET", "/badrequest", nil)
		w := httptest.NewRecorder()
		F(func(w http.ResponseWriter, r *http.Request) error {
			return OK(w, http.StatusOK, msg)
		}).ServeHTTP(w, req)
		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected http status ok, got %d", resp.StatusCode)
		}

		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}
		if !strings.Contains(string(bts), msg) {
			t.Fatalf("%q does not contain %q", string(bts), msg)
		}
	})
}

func TestHandlerFnNoError(t *testing.T) {
	req := httptest.NewRequest("GET", "/hf", nil)
	w := httptest.NewRecorder()
	HandlerFn(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}).ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected http status ok, got %d", resp.StatusCode)
	}
}

func TestHandlerFnWithError(t *testing.T) {
	req := httptest.NewRequest("GET", "/hf", nil)
	w := httptest.NewRecorder()
	status := http.StatusBadRequest
	msg := "data was wrong"
	HandlerFn(func(w http.ResponseWriter, r *http.Request) error {
		return Error{
			Status: status,
			Err:    fmt.Errorf(msg),
		}
	}).ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != status {
		t.Fatalf("expected http status %d, got %d", status, resp.StatusCode)
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if !strings.Contains(string(bts), msg) {
		t.Fatalf("%q does not contain %q", string(bts), msg)
	}
}

func TestHandlerFnWithUnknownError(t *testing.T) {
	req := httptest.NewRequest("GET", "/hf", nil)
	w := httptest.NewRecorder()
	status := http.StatusInternalServerError

	msg := "something was wrong"

	HandlerFn(func(w http.ResponseWriter, r *http.Request) error {
		return errors.New(msg)
	}).ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != status {
		t.Fatalf("expected http status %d, got %d", status, resp.StatusCode)
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if !strings.Contains(string(bts), msg) {
		t.Fatalf("%q does not contain %q", string(bts), msg)
	}
}