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
	"runtime"
	"strings"
	"time"
)

var env _env
var app_config configuration

//Holds local enviornmental information
type _env struct {
	Port       string
	Version    string
	Cmd_Server string
	OS         string
	Info       []string
}

type configuration struct {
	master_server     string
	port              string
	terminal_location string
	terminal_flags    string
	uuid              string
	registered        bool
}

//Saves the current configuration, over-writing what was present
func (self *configuration) save() error {
	j, _ := json.Marshal(self)
	return ioutil.WriteFile("app.confg", j, os.ModePerm)
}

//Message received from server
type message struct {
	Version  string
	Commands []Command
}

//Commands containing information on package management
type Command struct {
	Version     string
	Description string
	Cmd_type    string
	Order       int
	Exec        string
	Validates   []Validate_command
	Needed_file File
	Pass        bool
	Err         string
}

func (self *Command) Execute() error {
	var err error
	err = nil
	if self.Needed_file.Name != "" {
		err = download_file(self.Needed_file.Name,
			self.Needed_file.Url,
			self.Needed_file.Destination)
		if err != nil {
			self.Pass = false
			self.Err = err.Error()
			log.Printf("Error downloading file: %s", self.Err)
			return err
		}
	}
	var err_string string
	err, err_string = execute_command(self.Exec)
	if err != nil {
		self.Pass = false
		self.Err = err_string
	} else {
		self.Pass = true
		self.Err = ""
	}
	return err
}

//files to be downloaded
type File struct {
	Url         string
	Destination string
	Name        string
}

//Validations of successfully command completion
type Validate_command struct {
	Description     string
	Order           int
	Cmd_type        string
	Exec            string
	Expected_result string
	Pass            bool
	Err             string
}

func (self *Validate_command) Execute() error {
	err, err_string := execute_command(self.Exec)
	if err != nil {
		self.Err = err.Error() + ": " + err_string
		self.Pass = false
	} else {
		self.Err = ""
		self.Pass = true
	}
	return err
}

//Recieves commands from the server and kicks off the executing them
func cmd_handler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var cmds message
	err := decoder.Decode(&cmds)
	response := "OK"
	if err != nil {
		log.Println(err)
		response = err.Error()
		rw.Write([]byte(response))
	} else {
		log.Println(cmds.Version)
		rw.Write([]byte(response))
		go run_cmds(cmds.Commands)
	}
}

//Find the position of the cmd given a specified target order
func find_cmd_by_order(c []Command, target int) int {
	for i := 0; i < len(c); i++ {
		if c[i].Order == target {
			return i
		}
	}
	return -1
}

//Find the position of a given validator given a specified target order
func find_validator_by_order(c []Validate_command, target int) int {
	for i := 0; i < len(c); i++ {
		if c[i].Order == target {
			return i
		}
	}
	return -1
}

//Executes commands
func run_cmds(cmds []Command) {
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
			err := cmds[i].Execute()

			//set up validation
			v_order := 1
			v_length := len(cmds[i].Validates)
			v_validated := true
			//loop through validations running them in the correct order
			for v_index := 0; v_index < v_length && v_validated && err == nil; v_index++ {
				v_validated = false
				index := find_validator_by_order(cmds[i].Validates, v_order)
				if index > -1 {
					v_validated = true
					err = cmds[i].Validates[index].Execute()
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
	ex := exec.Command(app_config.terminal_location,
		app_config.terminal_flags,
		cmd)
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

func load_configuration() error {
	//Setup environmental info
	log.Println("Retrieving environmental information.")
	env.Info = os.Environ()
	env.OS = runtime.GOOS
	env.Port = app_config.port
	env.Version = "0.0.1"

	bytes, err := ioutil.ReadFile("app.confg")
	if err == nil {
		err = json.Unmarshal(bytes, &app_config)
		if err != nil {
			return err
		}
	}

	//set defaults if missing
	if env.OS == "linux" {
		if app_config.terminal_location == "" {
			app_config.terminal_location = "/bin/sh"
		}
		app_config.terminal_flags = "-c"
	} else {
		//pray we are on windows, for now
		if app_config.terminal_location == "" {
			app_config.terminal_location = "C:\\Windows\\System32\\cmd.exe"
		}
		app_config.terminal_flags = "/c"
	}

	if app_config.port == "" {
		app_config.port = ":8082"
		env.Port = app_config.port
	}
	return nil
}

//Checks in with the server
func check_in() {
	server := app_config.master_server + "/clients/checkin"
	post_json_to_server(server, "{\"status\": \"OK\", \"uuid\": \""+app_config.uuid+"\"}")
	time.Sleep(1 * time.Hour)
	check_in()
}

//Registers with the server
func register_with_server_if_needed() {
	if !app_config.registered {
		log.Println("Registering with server")
		server := app_config.master_server + "/clients/register"
		err := post_json_to_server(server, "body")
		if err != nil {
			app_config.registered = false
			log.Println("Failed to register with server: %s", err.Error())
			go retry_registering_with_server()
		} else {
			app_config.registered = true
			app_config.save()
		}
	}
}

//Retries to register with the server after a given time period
func retry_registering_with_server() {
	retry_in := 15 * time.Minute
	time.Sleep(retry_in)
	register_with_server_if_needed()
}

//Kicks off the program
func main() {

	log.Println("Loading configuration")
	err := load_configuration()
	if err != nil {
		log.Fatalf("Error loading configuration data: %s", err.Error())
	}

	register_with_server_if_needed()

	go check_in()

	//Set up server
	log.Println("Setting up server.")
	http.HandleFunc("/env", get_env)
	http.HandleFunc("/command", cmd_handler)
	log.Println("Starting server.")
	log.Println("Listening on port " + env.Port)
	log.Fatal(http.ListenAndServe(env.Port, nil))
}
