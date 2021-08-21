package core_test

import (
	"regexp"
	"testing"
)

func TestRegex(t *testing.T) {
	str := `time="2021-08-21T21:42:54Z" level=info msg="Block #13055460 finished (968ms)"`
	r := regexp.MustCompile(`time="(.*)"\s+level=(.*)\s+msg="(.*)"`)
	matches := r.FindStringSubmatch(str)
	if len(matches) != 4 {
		t.Fatalf("regexp.FindStringSubmatch failed : %v", matches)
	}
	if matches[1] != "2021-08-21T21:42:54Z" {
		t.Error("regexp.FindStringSubmatch failed")
	}
	if matches[2] != "info" {

		t.Error("regexp.FindStringSubmatch failed")
	}
	if matches[3] != "Block #13055460 finished (968ms)" {
		t.Error("regexp.FindStringSubmatch failed")
	}

}
