package fbauth

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
)

type mockGetter struct {
	err  error
	data string
}

type mockCloser struct{}

func (m mockCloser) Close() error {
	return nil
}

func (m mockGetter) Get(url string) (*http.Response, error) {
	type rc struct {
		io.Reader
		io.Closer
	}

	if m.err != nil {
		return nil, m.err
	}

	resp := http.Response{
		Body: rc{
			Reader: bytes.NewReader([]byte(m.data)),
			Closer: mockCloser{},
		},
	}
	return &resp, nil
}

func Test_Authenticate(t *testing.T) {
	type data struct {
		name  string
		id    string
		token string
		data  string
		err   error
	}

	tests := []data{
		data{name: "id is requird...", id: "", token: "foobar", err: errors.New("id is required")},
		data{name: "token is required...", id: "", token: "foobar", err: errors.New("token is required")},
		data{name: "bad data...", id: "foo", token: "bar", err: errors.New("failed to parse bad data"), data: `{"}`},
		data{name: "mismatching ids...", id: "foo", token: "bar", err: errors.New("facebook id mismatch :("), data: `{"id": "mismatched_id"}`},
		data{name: "passing test...", id: "foo", token: "bar", data: `{"id": "foo"}`},
	}

	for _, test := range tests {
		fmt.Println(test.name)
		mock := mockGetter{err: test.err, data: test.data}
		err := Authenticate(test.token, test.id, mock)

		if err != nil && test.err == nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if err == nil && test.err != nil {
			t.Errorf("expected an error: %v", err)
			return
		}
	}
}
