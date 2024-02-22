# C-Project

The C-Project web application provides a method of returning the last N number of lines from a log file via a
REST-like web interface.

## Getting Started

### Clone the Repo

```
git clone git@github.com:marklap/cproject.git
cd cproject
```

### Start the Server

In the `dist/` directory are pre-compiled executables. Choose the appropriate one for your
operating system.

> If you prefer to build the executable yourself, follow the
> [Build the Application](#build-the-application) instructions.

By default the server will listen on all interfaces (0.0.0.0) on TCP port 8080 and will only
respond to requests for log files in the `/var/log` directory.

To start the server using the defaults, execute the following command:

```
./bin/cproject
```

To start the server with a customized configuration, reference the [Command Line Usage](#command-line-usage) section or execute `./bin/project -h` to see the help information.

> **NOTE**: Ensure the server is started with a user account that has permissions to read the files that exist in 
> the directories provided with the `-prefixes` argument when starting the server.

### Tail a Log File

#### Requests

To tail a log file, make a POST http request with a JSON request body of the following structure.

**Structure**
```json
{
	"path": "/var/log/zoo.log",
	"num_lines": 10,
	"match_substrings": ["monkey", "octopus"],
	"case_sensitive": true,
}
```

Where:
- **path**: (*required*; string) the full path to a log file to tail
- **num_lines**: (integer) the number of lines to read from the end of the log file
- **match_substrings**: (list[string]) lines will only be returned if they match one of these strings
- **case_sensitive**: (boolean) if `match_substrings` is provided, set this to true to match in a case-sensitive manner

#### Responses

Responses are streamed ("chunked") to the client as they are yielded from the file reader. As such, while the content type of the response is technically text (`text/plain` MIME type), each chunk can be read as an individual JSON object.

**Structure**
```json
{"host": "web.server.zoo:8080", "line": "2024-02-20T07:10:42Z marklap fed the monkey 2 bananas"}
{"host": "web.server.zoo:8080", "line": "2024-02-20T07:10:42Z marklap fed the octopus 3 crabs"}
```

Where:
- **host:** (string) the host that responded with the `line`
- **line:** (string) a line from the log file

#### Examples

Using `curl`:

```
 $ curl -v localhost:8080/tail -d '{"num_lines":2,"path":"/var/log/zoo.log","match_substrings":["monkey","octopus"],"case_sensitive":true}'
*   Trying 127.0.0.1:8080...
* TCP_NODELAY set
* Connected to localhost (127.0.0.1) port 8080 (#0)
> POST /tail HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.68.0
> Accept: */*
> Content-Length: 102
> Content-Type: application/x-www-form-urlencoded
> 
* upload completely sent off: 102 out of 102 bytes
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Wed, 21 Feb 2024 15:13:40 GMT
< Content-Length: 1731
< Content-Type: text/plain; charset=utf-8
< 
{"host": "web.server.zoo:8080", "line": "2024-02-20T07:10:42Z marklap fed the monkey 2 bananas"}
{"host": "web.server.zoo:8080", "line": "2024-02-20T07:10:42Z marklap fed the octopus 3 crabs"}
* Connection #0 to host localhost left intact
```


## Appendix

### Command Line Usage
```
./bin/cproject -h
Usage of ./bin/cproject:
  -ip string
    	IP address to listen on (default "0.0.0.0")
  -port int
    	port to listen on (default 8080)
  -prefixes string
    	path prefixes to use for path validation [':' deliminted] (default "/var/log")
```

### Build the Application

Ensure you have [go 1.20+](https://go.dev/dl/) installed.

To build using Make:
```
make build
```

To build for *nix/Mac using the go build tool:
```
go build -o ./bin/cproject ./cmd/
```

To build for Windows using the go build tool:
```
go build -o ./bin/cproject.exe ./cmd/
```

### Running Tests

Run tests using Make:

```
make test
```

Run tests using go test tool:

```
go test -v
```