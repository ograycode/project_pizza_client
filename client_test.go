package main

import (
	"runtime"
	"testing"
)

//Tests the ability to create and download files through
//the download_file function
func Test_download_file(t *testing.T) {
	err := download_file("gopher.png",
		"http://golang.org/doc/gopher/frontpage.png",
		"test")
	if err != nil {
		t.Errorf("Err: %s", err.Error())
	}
}

//Tests loading the configuration
func Test_load_configuration(t *testing.T) {
	err := load_configuration()
	if err != nil {
		t.Logf("Failed to load configuration: %s", err)
		t.Logf("app_config.terminal_location: %s", app_config.terminal_location)
		t.Logf("app_config.terminal_flags: %s", app_config.terminal_flags)
		t.Logf("app_config.port: %s", app_config.port)
		t.Logf("app_config.master_server: %s", app_config.master_server)
	}
}

//Tests validates execution function
func Test_Validate(t *testing.T) {

	exec := ""

	if runtime.GOOS == "linux" {
		exec = "touch test/validator.test"
	} else {
		exec = "echo \"\" > test/validator.test"
	}

	var val validate_command
	val.description = "Test command"
	val.exec = exec
	val.err = ""
	val.pass = false

	val.execute()

	if !val.pass {
		t.Errorf("Err: %s", val.err)
	}
}

//Tests Command's executrion function
func Test_Cmd(t *testing.T) {

	exec := ""

	if runtime.GOOS == "linux" {
		exec = "touch test/cmd.test"
	} else {
		exec = "echo \"\" > test/cmd.test"
	}

	var cmd command
	cmd.description = "Test"
	cmd.exec = exec
	cmd.err = ""
	cmd.pass = false

	cmd.execute()

	if !cmd.pass {
		t.Errorf("Command err: %s", cmd.err)
	}
}

//Tests find_validator_by_order as well find_cmd_by_order
func Test_find_by_order(t *testing.T) {
	//Start testing find_validator_by_order
	var val1 validate_command
	var val2 validate_command
	var val3 validate_command

	val1.order = 1
	val2.order = 2
	val3.order = 3

	vals := []validate_command{val1, val2, val3}

	expected_position := 1
	position := find_validator_by_order(vals, 2)
	if expected_position != position {
		t.Errorf("Expected %i got %i", expected_position, position)
	}

	expected_position = -1
	position = find_validator_by_order(vals, 5)
	if expected_position != position {
		t.Errorf("Expected %i got %i", expected_position, position)
	}

	//Start  testing find_cmd_by_order
	var cmd1 command
	var cmd2 command
	var cmd3 command

	cmd1.order = 1
	cmd2.order = 5
	cmd3.order = 3

	cmds := []command{cmd1, cmd2, cmd3}

	expected_position = 2
	position = find_cmd_by_order(cmds, 3)
	if expected_position != position {
		t.Errorf("Expected %i got %i", expected_position, position)
	}

	expected_position = -1
	position = find_cmd_by_order(cmds, 4)
	if expected_position != position {
		t.Errorf("Expected %i got %i", expected_position, position)
	}

}
