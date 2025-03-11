package bind

import (
	"errors"
	"io"
	"net/http"
)

type reader func(r *http.Request) (string, error)

func makeQueryReader(name string, req bool, def string) reader {
	return func(r *http.Request) (string, error) {
		var val string
		if r.URL.Query().Has(name) {
			val = r.URL.Query()[name][0] // TODO lists
		}
		if req && val == "" {
			return "", errors.New("'" + name + "' required and not given")
		}
		if val == "" {
			val = def
		}
		return val, nil
	}
}

func makePathReader(name string) reader {
	return func(r *http.Request) (string, error) {
		return r.PathValue(name), nil
	}
}

func bodyReader(r *http.Request) (string, error) {
	if bytes, err := io.ReadAll(r.Body); err != nil {
		return "", err
	} else {
		return string(bytes), nil
	}
}
