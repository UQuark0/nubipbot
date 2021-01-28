package nubip

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type NubipAPI struct {
	client *http.Client
}

const nubipHost = "https://nubip.dots.org.ua/"

func NewNubipAPI(username, password string) (*NubipAPI, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	n := &NubipAPI{
		client: &http.Client{
			Jar: jar,
		},
	}

	err = n.authenticate(username, password)
	if err != nil {
		return nil, err
	}
	return n, nil
}

var (
	errInvalidCredentials = errors.New("Пользователь с таким именем и паролем не найден")
)

func (n *NubipAPI) authenticate(username, password string) error {
	data := url.Values{
		"username": {username},
		"password": {password},
	}

	resp, err := n.client.PostForm(nubipHost+"login", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(text), "Пользователь с таким именем и паролем не найден") {
		return errInvalidCredentials
	}

	return nil
}
