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

	var val Validate_command
	val.Description = "Test command"
	val.Exec = exec
	val.Err = ""
	val.Pass = false

	val.Execute()

	if !val.Pass {
		t.Errorf("Err: %s", val.Err)
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

	var cmd Command
	cmd.Description = "Test"
	cmd.Exec = exec
	cmd.Err = ""
	cmd.Pass = false

	cmd.Execute()

	if !cmd.Pass {
		t.Errorf("Command err: %s", cmd.Err)
	}
}

//Tests find_validator_by_order as well find_cmd_by_order
func Test_find_by_order(t *testing.T) {
	//Start testing find_validator_by_order
	var val1 Validate_command
	var val2 Validate_command
	var val3 Validate_command

	val1.Order = 1
	val2.Order = 2
	val3.Order = 3

	vals := []Validate_command{val1, val2, val3}

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
	var cmd1 Command
	var cmd2 Command
	var cmd3 Command

	cmd1.Order = 1
	cmd2.Order = 5
	cmd3.Order = 3

	cmds := []Command{cmd1, cmd2, cmd3}

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
