package forms

import "net/url"

type Form struct {
	Values url.Values
	Errors errors
}

func New(data url.Values) *Form {
	return &Form{
		url.Values(data),
		errors(map[string][]string{}),
	}
}
