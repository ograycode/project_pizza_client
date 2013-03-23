# Project Pizza #

Project Pizza is a simple way to remotely execute commands on any box without additional dependencies. 

## Client ##
The client is responsible for receiving commands from the master server, executing them and then telling the master server their status.

In addition, the client is also responsible for periodically letting the server know its' own status, as well as the the machine's environmental information. This servers a few purposes: 1) allow for auditing of machines; and 2) to always allow you to know which machines are currently running.

### Why Go ###
The client is written in [Go] (http://www.golang.org "Go") to ensure a wide-range of cross compatibility as well as to limit the number of dependencies that a system must have down to one -- the executable itself. Run a Rails server? That's fine, write your scripts in Ruby. Maintain a Windows box? Write them in C#.

## Server TBD ##

## License ##
Project pizza is dual licensed under MIT and GPLv3