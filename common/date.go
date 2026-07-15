package common

import (
	"fmt"
	"strings"
	"time"
)

type JDate time.Time

const layout = "2006-01-02 15:04:05"

var nilTime = (time.Time{}).UnixNano()

func (d *JDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		*d = JDate{}
		return
	}
	t, err := time.Parse(layout, s)
	if err != nil {
		// dhtmlxGantt 10+ DataProcessor sends dates in ISO format
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return err
		}
		t = t.UTC()
	}
	*d = JDate(t)

	return
}

func (d *JDate) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}
	if time.Time(*d).UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", time.Time(*d).Format(layout))), nil
}

func (d *JDate) IsEmpty() bool {
	return d == nil || time.Time(*d).UnixNano() == nilTime
}
