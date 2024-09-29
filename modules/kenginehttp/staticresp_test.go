package kenginehttp

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/khulnasoft/kengine"
)

func TestStaticResponseHandler(t *testing.T) {
	r := fakeRequest()
	w := httptest.NewRecorder()

	s := Static{
		StatusCode: http.StatusNotFound,
		Headers: http.Header{
			"X-Test": []string{"Testing"},
		},
		Body:  "Text",
		Close: true,
	}

	err := s.ServeHTTP(w, r)
	if err != nil {
		t.Errorf("did not expect an error, but got: %v", err)
	}

	resp := w.Result()
	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d but got %d", http.StatusNotFound, resp.StatusCode)
	}
	if resp.Header.Get("X-Test") != "Testing" {
		t.Errorf("expected x-test header to be 'testing' but was '%s'", resp.Header.Get("X-Test"))
	}
	if string(respBody) != "Text" {
		t.Errorf("expected body to be 'test' but was '%s'", respBody)
	}
}

func fakeRequest() *http.Request {
	r, _ := http.NewRequest("GET", "/", nil)
	repl := kengine.NewReplacer()
	ctx := context.WithValue(r.Context(), kengine.ReplacerCtxKey, repl)
	r = r.WithContext(ctx)
	return r
}
