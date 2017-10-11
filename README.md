# LogRouter

Small app to read stdin and send data onto specified logging location.  The normal use case is to pipe the output (stdout/stderr) of an application to LogRouter.

Currently the only supported logging destination is Graylog via a GELF UDP input.

## Usage

Command line flags are used to pass in parameters.

| Flag               | Description                                                                                           | Default Value |
| ------------------ | ----------------------------------------------------------------------------------------------------- | ------------- |
| input-format       | Defines the expected input format to receive.  Currently one of logfmt, json or unknown is supported. | none          |
| output             | Defines the output destination.  Currently only Graylog is supported.                                 | none          |
| graylog-address    | IP address or hostname of the Graylog server to send logs to.                                         | none          |
| graylog-port       | UDP GELF port.                                                                                        | none          |
| graylog-attributes | Comma separated list of attributes to pass to Graylog in the form name:value.                         | none          |
| debug              | Enable debug logging.                                                                                 | false         |
| version            | Display version and build number                                                                      | false         |

The below redirects stderr (2) from do_something.exe to stdout (1) and then pipes stdout to logrouter.exe.  LogRouter will then send the message to the Graylog server at 192.168.1.100 on port 12202.  It will also attach the specified attributes to the log message.

````Batchfile
do_something.exe 2>&1 | logrouter.exe --input-format logfmt --output graylog --graylog-address 192.168.1.100 --graylog-port 12202 --attributes "application:do_something,environment:prod"
````

## Downloading a release

<https://github.com/rokett/LogRouter/releases>

## Building the executable

All dependencies are version controlled, so building the project is really easy.

1. Clone the repository locally.
2. From within the repository directory run `go build`.
3. Hey presto, you have an executable.
