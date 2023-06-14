package url

import (
	gourl "net/url"
)

type URL struct {
	*gourl.URL
}

func (u *URL) Equal(b *URL) bool {
	return u.String() == b.String()
}

func (u *URL) MarshalText() (text []byte, err error) { //nolint: unparam
	return []byte(u.String()), nil
}

func (u *URL) UnmarshalText(text []byte) error {
	u1, err := gourl.Parse(string(text))
	if err != nil {
		return err
	}
	u.URL = u1
	return nil
}
