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

![Evans](./evans.gif)  

## Motivation
Evans was created to use easier than other existing gRPC clients.  
If you want to keep your product quality, you must use CI with gRPC testing, should not do use manually testing.  
Evans will complete your other usecases just like:  

- to automate some tasks by scripting
- manullary gRPC API inspection

above usecases is corresponding to Evans's two modes, command-line mode and REPL mode.  

## command-line mode
command-line mode is stateless mode just like [grpc-ecosystem/polyglot](https://github.com/grpc-ecosystem/polyglot).  
command-line mode issue one request per one command as its name suggests.  
So it is based on UNIX philosophy.  

For example, read inputs from `stdin`, the command will be a filter command.  
On the other hand, the command result will be outputted to `stdout` by JSON formatted.  
So, you can format it by any commands like `jq`.  

``` sh
$ echo '{ "name": "ktr" }' | evans --package hello --service Greeter --call SayHello hello.proto | jq -r '.message'
# hello, ktr!
```

Also, if you want to use same command (e.g. use same JSON inputs), you can use `--file` (`-f`) option.  

``` sh
$ cat hello.json
{
  "name": "ktr"
}

$ evans -f hello.json --package hello --service Greeter --call SayHello hello.proto | jq -r '.message'
# hello, ktr!
```

By the way, command-line mode is not able to omit `--package`, `--service` and `--call` option unlike REPL mode.  
However, if `.evans.toml` is exist in Git project root, you can denote default values.  

``` toml
[default]
package = "hello"
service = "Greeter"
```

Then, command will be more clearly.  

``` sh
$ evans -f hello.json --call SayHello hello.proto
```

## REPL mode
REPL mode is the solution for second usecase.  
You can use it without thinking like package name, service name, RPC name, command usage, and so on because REPL mode has powerful completion!  

Actual demonstration:
![demo](./evans.gif)  

proto file which read in demonstration is:  
``` proto
syntax = "proto3";

package helloworld;

service Greeter {
  rpc SayHello (HelloRequest) returns (HelloResponse) {}
}

enum Language {
  ENGLISH = 0;
  JAPANESE = 1;
}

message Person {
  string name = 1;
  repeated string others = 2;
}

message HelloRequest {
  Person person = 1;
  Language language = 2;
}

message HelloResponse {
  string message = 1;
}
```

implementation of the server is:
``` go
package main

import (
    "fmt"
    "log"
    "net"
    "strings"

    "golang.org/x/net/context"

    helloworld "github.com/ktr0731/evans/tmp"
    "google.golang.org/grpc"
)

type Greeter struct{}

func (t *Greeter) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloResponse, error) {
    msg := "Hello, %s also %s!"
    if req.GetLanguage() == helloworld.Language_JAPANESE {
        msg = "こんにちは、%s と %s ！"
    }
    return &helloworld.HelloResponse{
        Message: fmt.Sprintf(msg, req.GetPerson().GetName(), strings.Join(req.GetPerson().GetOthers(), ", ")),
    }, nil
}

func main() {
    l, err := net.Listen("tcp", "localhost:50051")
    if err != nil {
        log.Fatal(err)
    }

    server := grpc.NewServer()
    helloworld.RegisterGreeterServer(server, &Greeter{})
    if err := server.Serve(l); err != nil {
        log.Fatal(err)
    }
}
```

## Installation
### from binary
please see [GitHub Releases](https://github.com/ktr0731/evans/releases).  

### macOS
``` sh
$ brew tap ktr0731/evans
$ brew install evans
```

### go get
``` sh
$ go get github.com/ktr0731/evans
```


## Usage
Evans consists of some commands in REPL.  

Enter to REPL (this file is [here](./testdata/proto/helloworld/helloworld.proto))  
``` 
$ evans testdata/proto/helloworld/helloworld.proto
```

To show summary of packages, services or messages of proto files REPL read:  
``` 
> show package
+------------+
|  PACKAGE   |
+------------+
| helloworld |
+------------+

> show service
+---------+----------+--------------+--------------+
| SERVICE |   RPC    | REQUESTTYPE  | RESPONSETYPE |
+---------+----------+--------------+--------------+
| Greeter | SayHello | HelloRequest | HelloReply   |
+---------+----------+--------------+--------------+

> show message
+--------------+
|   MESSAGE    |
+--------------+
| HelloRequest |
| HelloReply   |
+--------------+
```

To show more description of message:  
``` 
> desc HelloRequest
+-------+-------------+
| FIELD |    TYPE     |
+-------+-------------+
| name  | TYPE_STRING |
+-------+-------------+
```

Call a RPC:  
``` 
> call SayHello
name (TYPE_STRING) = ktr
{
  "message": "Hello ktr"
}
```

Evans constructs a gRPC request interactively and send the request to gRPC server.  
Finally, Evans prints the JSON formatted result.  

## See Also
Evans (DJ YOSHITAKA)  
![Evans](./evans.png)  
[iTunes](https://itunes.apple.com/jp/album/jubeat-original-soundtrack/id325295989)  
