# bibService
A small web service to provide information about BIB records.

## Source code
The main source code files are:

* `main.go`: The launcher
* `server.go`: The web server router (i.e. the controllers)
* `sierra/`: Code to interface with the Sierra API
* `bibModel/`: Code to handle bib records level operations.

## Running the service
To run the service you'll need a `settings.json` file with the configuration information to connect to Sierra. Take a look at the `settings.sample.json` for an example. Once you have a `settings.json` file created you just need to run:

```
$ go build
$ ./bibService settings.json
```

The service will be listening for requests at the `serverAddress` indicated in `settings.json`. You can test is with a command like this:

```
$ curl localhost:9001/status
OK
```


## Deploying the service
To deploy the service to a Linux server:

```
$ GOOS=linux go build
$ scp bibService the-server:./the-path/
```

You'll need to create a `settings.json` file on the server too.
