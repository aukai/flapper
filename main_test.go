package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestHandlers(t *testing.T) {
	t.Run("publishes environment variables", func(t *testing.T) {
		os.Setenv("O_VARIABLE_1", "bird")
		os.Setenv("X_VARIABLE_1", "dolphin")

		type variable struct {
			name    string
			enabled bool
		}

		var expected = []variable{
			{
				name:    "bird",
				enabled: true,
			},
			{
				name:    "dolphin",
				enabled: false,
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/env", nil)
		w := httptest.NewRecorder()
		publishEnvVars(w, req)
		res := w.Result()
		defer res.Body.Close()
		var data []variable
		err := json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			t.Errorf("expected error to be nil go %v", err)
		}
		if reflect.DeepEqual(data, expected) {
			t.Errorf("expected %v got %v", expected, data)
		}
	})

	t.Run("publishes version", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/version", nil)
		w := httptest.NewRecorder()
		publishEnvVars(w, req)
		res := w.Result()
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("expected error to be nil go %v", err)
		}
		if strings.Contains(string(data), "flapper_version") {
			t.Errorf("expected %v to contain flapper_version", data)
		}
	})
}
