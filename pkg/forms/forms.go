package forms

import "net/url"

type Form struct {
	Testget string
	Values  url.Values
	Errors  errors
}

func New(data url.Values) *Form {
	return &Form{
		"kek",
		url.Values(data),
		errors(map[string][]string{}),
	}
}

func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
