package fbauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// URLGetter allows for methods such as http.Get, urlfetch.Get methods to be used
// to make the external http request to perform the facebook authentication
type URLGetter interface {
	Get(url string) (*http.Response, error)
}

// Authenticate authenticates the facebook id and returns an error on failure
func Authenticate(fbToken, fbID string, client URLGetter) error {
	var fbp map[string]interface{}
	_, err := fetch(client, fbToken, []string{"id"}, &fbp)
	if err != nil {
		return err
	}
	if fbp["id"] != fbID {
		return fmt.Errorf("facebook id mismatch, %v != %v", fbp["id"], fbID)
	}

	return nil
}

// GetPhoto obtains the content-type and photo closest to the specified width
func GetPhoto(token string, width int, getter URLGetter) ([]byte, string, error) {
	type fbImageResp struct {
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	var payload fbImageResp
	_, err := fetch(getter, token, []string{fmt.Sprintf("picture.width(%d)", width)}, &payload)
	if err != nil {
		return nil, "", err
	}

	resp, err := getter.Get(payload.Picture.Data.URL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch facebook profile image: %v", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return data, resp.Header.Get("Content-Type"), nil
}

func makeURL(token string, fields []string) (*url.URL, error) {
	u, err := url.Parse("https://graph.facebook.com/v2.7/me")
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %v", err)
	}
	q := u.Query()
	q.Set("access_token", token)
	q.Set("fields", strings.Join(fields, "&"))
	u.RawQuery = q.Encode()

	return u, nil
}

func fetch(getter URLGetter, token string, fields []string, dest interface{}) (map[string][]string, error) {
	if len(token) == 0 {
		return nil, errors.New("missing facebook token")
	}
	if len(fields) == 0 {
		return nil, errors.New("missing fields")
	}

	url, err := makeURL(token, fields)
	if err != nil {
		return nil, err
	}
	resp, err := getter.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(dest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode fb auth response: %v", err)
	}

	return map[string][]string(resp.Header), nil
}
