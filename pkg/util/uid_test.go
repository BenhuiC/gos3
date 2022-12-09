package util

import (
	"regexp"
	"testing"
)

func TestUUID32(t *testing.T) {
	id := UUID32()
	if match, _ := regexp.Match("^[0-9a-zA-Z]{1,32}$", []byte(id)); !match {
		t.Fail()
	}
}
