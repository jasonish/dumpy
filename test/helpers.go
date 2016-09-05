package test

import "testing"

func FailIf(t *testing.T, value bool) {
	if value {
		t.FailNow()
	}
}

func FailIfNot(t *testing.T, value bool) {
	FailIf(t, !value)
}

func FailIfNil(t *testing.T, value interface{}) {
	if value == nil {
		t.FailNow()
	}
}