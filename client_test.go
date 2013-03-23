package main

import (
	"testing"
)

func Test_Validate(t *testing.T) {
	var val validate_command
	val.description = "Test command"
	val.exec = "touch foo.test"
	val.err = ""
	val.pass = false

	val.execute()

	if !val.pass {
		t.Errorf("Err: %s", val.err)
		t.Fail()
	}
}

func Test_download_file(t *testing.T) {
	err := download_file("gopher.png",
		"http://golang.org/doc/gopher/frontpage.png",
		"test")
	if err != nil {
		t.Errorf("Err: %s", err.Error())
		t.Fail()
	}
}
