package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var env _env

//Holds local enviornmental information
type _env struct {
	Port       string
	Version    string
	Cmd_Server string
	Info       []string
}

//Message received from server
type message struct {
	versoin  string
	commands []command
}

//Commands containing information on package management
type command struct {
	version     string
	description string
	cmd_type    string
	order       int
	exec        string
	validates   []validate_command
	needed_file file
	pass        bool
	err         string
}

func (self *command) execute() error {
	var err error
	err = nil
	if self.needed_file.name != "" {
		err = download_file(self.needed_file.name,
			self.needed_file.url,
			self.needed_file.destination)
		if err != nil {
			self.pass = false
			self.err = err.Error()
			log.Printf("Error downloading file: %s", self.err)
			return err
		}
	}
	var err_string string
	err, err_string = execute_command(self.exec)
	if err != nil {
		self.pass = false
		self.err = err_string
	} else {
		self.pass = true
		self.err = ""
	}
	return err
}

//files to be downloaded
type file struct {
	url         string
	destination string
	name        string
}

//Validations of successfully command completion
type validate_command struct {
	description     string
	order           int
	cmd_type        string
	exec            string
	expected_result string
	pass            bool
	err             string
}

func (self *validate_command) execute() error {
	err, err_string := execute_command(self.exec)
	if err != nil {
		self.err = err.Error() + ": " + err_string
		self.pass = false
	} else {
		self.err = ""
		self.pass = true
	}
	return err
}

//Recieves commands from the server and kicks off the executing them
func cmd_handler(rw http.ResponseWriter, req *http.Request) {
	s_cmds := req.FormValue("c")
	cmds := message{}
	err := json.Unmarshal([]byte(s_cmds), &cmds)
	if err != nil {
		log.Println(err)
	} else {
		run_cmds(cmds.commands)
	}
}

//Find the position of the cmd given a specified target order
func find_cmd_by_order(c []command, target int) int {
	for i := 0; i < len(c); i++ {
		if c[i].order == target {
			return i
		}
	}
	return -1
}

//Find the position of a given validator given a specified target order
func find_validator_by_order(c []validate_command, target int) int {
	for i := 0; i < len(c); i++ {
		if c[i].order == target {
			return i
		}
	}
	return -1
}

//Executes commands
func run_cmds(cmds []command) {
	order := 1
	executed := true
	var err error
	err = nil

	//loop through the commands executing and validating
	//TODO: Optimize, because this is super-slow
	for j := 0; j < len(cmds) && executed && err == nil; j++ {
		executed = false
		i := find_cmd_by_order(cmds, order)
		if i > -1 {
			executed = true
			err := cmds[i].execute()

			//set up validation
			v_order := 1
			v_length := len(cmds[i].validates)
			v_validated := true
			//loop through validations running them in the correct order
			for v_index := 0; v_index < v_length && v_validated && err == nil; v_index++ {
				v_validated = false
				index := find_validator_by_order(cmds[i].validates, v_order)
				if index > -1 {
					v_validated = true
					err = cmds[i].validates[index].execute()
				}
				v_order++
			}
		}
		order++
	}
	m := message{"", cmds}
	post_message_to_server("url", m)
}

//Sends a pre-formatted message to teh server
func post_json_to_server(url string, body string) error {
	r_body := strings.NewReader(body)
	http.Post("url", "json", r_body)
	return nil
}

//Sends a report to the server regarding executing commands
func post_message_to_server(url string, msg message) error {
	j_msg, _ := json.Marshal(msg)
	http.Post("url", "json", bytes.NewReader(j_msg))
	return nil
}

//Download a file
//returns quickly on any error
func download_file(file_name string, url string, destination string) error {

	//get our starting directory
	starting_dir, erro := os.Getwd()
	if erro != nil {
		return erro
	}

	//if we have a destination, create and cd into it
	if destination != "" {
		err := os.MkdirAll(destination, os.ModePerm)

		if err != nil {
			return err
		}

		err = os.Chdir(destination)
		if err != nil {
			return err
		}
	}

	//create our placeholder file
	out, create_err := os.Create(file_name)
	defer out.Close()

	if create_err != nil {
		return create_err
	}

	//download the file
	resp, download_err := http.Get(url)
	defer resp.Body.Close()

	if download_err != nil {
		return download_err
	}

	//copy it onto the hard-disk
	_, copy_err := io.Copy(out, resp.Body)

	if copy_err != nil {
		return copy_err
	}

	//change back to the starting directory if needed
	curr_dir, er := os.Getwd()
	if er != nil {
		return er
	} else if curr_dir != starting_dir {
		os.Chdir(starting_dir)
	}

	return nil
}

//execute command
func execute_command(cmd string) (error, string) {
	var error_string string
	error_string = ""
	ex := exec.Command("/bin/sh", "-c", cmd)
	//pipe stderr
	stderr, err := ex.StderrPipe()
	if err != nil {
		return err, "Error piping stderr"
	}

	err = ex.Start()
	//get stderr
	b, b_err := ioutil.ReadAll(stderr)
	if b_err != nil {
		error_string = "Error getting stderr message: " + b_err.Error()
	} else {
		error_string = "STDERR: " + string(b)
	}

	err = ex.Wait()
	if err != nil {
		return err, error_string
	}
	return err, ""
}

//Returns environmental information to those who query
func get_env(rw http.ResponseWriter, req *http.Request) {
	env_bytes, err := json.Marshal(env)
	if err != nil {
	} else {
		rw.Write(env_bytes)
	}
}

//Kicks off the program
func main() {
	//Setup environmental info
	log.Println("Retrieving environmental information.")
	env.Port = ":8082"
	env.Version = "0.0.1"
	env.Info = os.Environ()

	//Set up server
	log.Println("Setting up server.")
	http.HandleFunc("/env", get_env)
	http.HandleFunc("/command", cmd_handler)
	log.Println("Starting server.")
	log.Fatal(http.ListenAndServe(env.Port, nil))
}
