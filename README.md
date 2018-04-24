```
  ______
 |  ____|
 | |__    __   __   __ _   _ __    ___
 |  __|   \ \ / /  / _. | | '_ \  / __|
 | |____   \ V /  | (_| | | | | | \__ \
 |______|   \_/    \__,_| |_| |_| |___/

 more expressive universal gRPC client
```
--- 

[![CircleCI](https://circleci.com/gh/ktr0731/evans/tree/master.svg?style=svg)](https://circleci.com/gh/ktr0731/evans/tree/master)  

## Motivation
Evans was created to use easier than other existing gRPC clients.  
If you want to keep your product quality, you must use CI with gRPC testing, should not do use manually testing.  
Evans will complete your other usecases just like:  

- manually gRPC API inspection
- to automate some tasks by scripting

above usecases is corresponding to Evans's two modes, REPL mode, and command-line mode.  

## REPL mode
![Evans](./evans1.gif)  
REPL mode is the solution for first usecase.  
You can use it without thinking like package name, service name, RPC name, command usage, and so on because REPL mode has powerful completion!  

proto file which read in demonstration and its implementation are available at [ktr0731/evans-demo](https://github.com/ktr0731/evans-demo).  

## command-line mode
![Evans](./evans2.gif)  

command-line mode is stateless mode just like [grpc-ecosystem/polyglot](https://github.com/grpc-ecosystem/polyglot).  
command-line mode issue one request per one command as its name suggests.  
So it is based on UNIX philosophy.  

For example, read inputs from `stdin`, the command will be a filter command.  
On the other hand, the command result will be outputted to `stdout` by JSON formatted.  
So, you can format it by any commands like `jq`. Also, if you want to use the same command (e.g. use same JSON inputs), you can use `--file` (`-f`) option.  

By the way, command-line mode is not able to omit `--package`, `--service` and `--call` option, unlike REPL mode.  
However, if `.evans.toml` is exist in Git project root, you can denote default values.  

``` toml
[default]
protoFile = ["api/api.proto"]
package = "api"
service = "UserService"
```

Then, the command will be more clear.  

## Installation
highly recommended methods are GitHub Releases or HomeBrew because these can be software update automatically by the built-in feature in Evans.  

### from GitHub Releases
please see [GitHub Releases](https://github.com/ktr0731/evans/releases).  

### macOS
``` sh
$ brew tap ktr0731/evans
$ brew install evans
```

### go get
v1.10 or later required.  
``` sh
$ go get github.com/ktr0731/evans
```

## Usage
Evans consists of some commands in REPL.  

Enter to REPL (this file is [here](adapter/gateway/testdata/helloworld.proto))  
``` 
$ evans adapter/gateway/testdata/helloworld.proto
```

To show the summary of packages, services or messages of proto files REPL read:  
``` 
> show package
+------------+
|  PACKAGE   |
+------------+
| helloworld |
+------------+

> show service
+---------+----------+--------------+---------------+
| SERVICE |   RPC    | REQUESTTYPE  | RESPONSETYPE  |
+---------+----------+--------------+---------------+
| Greeter | SayHello | HelloRequest | HelloResponse |
+---------+----------+--------------+---------------+

> show message
+---------------+
|    MESSAGE    |
+---------------+
| HelloRequest  |
| HelloResponse |
+---------------+
```

To show more description of a message:  
``` 
> desc HelloRequest
+---------+-------------+
|  FIELD  |    TYPE     |
+---------+-------------+
| name    | TYPE_STRING |
| message | TYPE_STRING |
+---------+-------------+
```

Set headers for each request:
```
> header foo=bar
```

To show headers:
```
> show header
+------------+-------+
|    KEY     |  VAL  |
+------------+-------+
| user-agent | evans |
| foo        | bar   |
+------------+-------+
```

Call a RPC:  
``` 
> call SayHello
name (TYPE_STRING) = ktr
message (TYPE_STRING) => hello!
```

Evans constructs a gRPC request interactively and sends the request to a gRPC server.  
Finally, Evans prints the JSON formatted result.  

## Supported IDL (interface definition language)
- [Protocol Buffers 3](https://developers.google.com/protocol-buffers/)  

## See Also
Evans (DJ YOSHITAKA)  
![Evans](./evans.png)  
[iTunes](https://itunes.apple.com/jp/album/jubeat-original-soundtrack/id325295989)  
