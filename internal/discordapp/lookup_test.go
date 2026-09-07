package discordapp

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// doerFunc adapts a function to HTTPDoer.
type doerFunc func(*http.Request) (*http.Response, error)

func (d doerFunc) Do(req *http.Request) (*http.Response, error) { return d(req) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestLookup_Name_ReturnsNameOnSuccess(t *testing.T) {
	var gotURL string
	doer := doerFunc(func(req *http.Request) (*http.Response, error) {
		gotURL = req.URL.String()
		return jsonResponse(http.StatusOK, `{"id":"123","name":"Jungle diff"}`), nil
	})

	got, err := New(doer).Name(context.Background(), "123")
	if err != nil {
		t.Fatalf("Name: %v", err)
	}
	if got != "Jungle diff" {
		t.Errorf("Name() = %q, want %q", got, "Jungle diff")
	}
	if want := "https://discord.com/api/v10/applications/123/rpc"; gotURL != want {
		t.Errorf("requested URL = %q, want %q", gotURL, want)
	}
}

func TestLookup_Name_ErrorsOnNonOKStatus(t *testing.T) {
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusNotFound, `{"message":"Unknown Application"}`), nil
	})

	if _, err := New(doer).Name(context.Background(), "bogus"); err == nil {
		t.Fatal("Name() did not error on a 404")
	}
}

func TestLookup_Name_ErrorsOnEmptyName(t *testing.T) {
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"id":"123"}`), nil
	})

	if _, err := New(doer).Name(context.Background(), "123"); err == nil {
		t.Fatal("Name() did not error on a missing name field")
	}
}

func TestLookup_Name_ErrorsOnTransportFailure(t *testing.T) {
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})

	if _, err := New(doer).Name(context.Background(), "123"); err == nil {
		t.Fatal("Name() did not propagate a transport error")
	}
}
