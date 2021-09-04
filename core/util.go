package core

import (
	"regexp"
	"time"
)

func arrayContains(array []string, value string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

type Params struct {
	Time  time.Time
	Level string
	Msg   string
}

func extractParams(line string) (*Params, error) {
	r := regexp.MustCompile(`time="(.*)"\s+level=(.*)\s+msg="(.*)"`)
	matches := r.FindStringSubmatch(line)
	if len(matches) == 0 {
		return nil, errNoMatches
	}

	t, err := time.Parse("2006-01-02T15:04:05Z", matches[1])
	if err != nil {
		return nil, err
	}

	return &Params{
		Time:  t,
		Level: matches[2],
		Msg:   matches[3],
	}, nil

}
