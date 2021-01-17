![Evans](https://user-images.githubusercontent.com/12953836/53423552-e5ca8800-3a24-11e9-9927-fe7f3d5f867a.png)

--- 

[![GitHub Actions](https://github.com/ktr0731/evans/workflows/main/badge.svg)](https://github.com/ktr0731/evans/actions)
[![codecov](https://codecov.io/gh/ktr0731/evans/branch/master/graph/badge.svg)](https://codecov.io/gh/ktr0731/evans)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ktr0731/evans)](https://pkg.go.dev/github.com/ktr0731/evans)

## Motivation
Evans has been created to use easier than other existing gRPC clients.  
If you want to keep your product quality, you must use CI with gRPC testing, should not do use manual testing.  
Evans will complete your other use cases just like:  

- Manually gRPC API inspection
- To automate some tasks by scripting

The above use cases are corresponding to Evans's two modes, REPL mode, and CLI mode.  

## REPL mode
![Evans](./evans1.gif)  
REPL mode is the solution for first use case.  
You can use it without thinking like the package name, service name, RPC name, command usage, and so on because REPL mode has powerful completion!  

## CLI mode
![Evans](./evans2.gif)  

CLI mode is a stateless mode just like [grpc-ecosystem/polyglot](https://github.com/grpc-ecosystem/polyglot).  
It sends one request per one command as its name suggests.  
So it is based on UNIX philosophy.  

For example, read inputs from `stdin`, the command will be a filter command.  
On the other hand, the command result will be outputted to `stdout` by JSON formatted.  
So, you can format it by any commands like `jq`. Also, if you want to use the same command (e.g. use same JSON inputs), you can use `--file` (`-f`) option.  

## Table of Contents
- [Installation](#installation)
   - [From GitHub Releases](#from-github-releases)
   - [macOS](#macos)
   - [[Not-recommended] go get](#not-recommended-go-get)
- [Usage (REPL)](#usage-repl)
   - [Basic usage](#basic-usage)
   - [Repeated fields](#repeated-fields)
   - [Enum fields](#enum-fields)
   - [Bytes type fields](#bytes-type-fields)
   - [Client streaming RPC](#client-streaming-rpc)
   - [Server streaming RPC](#server-streaming-rpc)
   - [Bidirectional streaming RPC](#bidirectional-streaming-rpc)
   - [Skip the rest of fields](#skip-the-rest-of-fields)
   - [Enriched response](#enriched-response)
- [Usage (CLI)](#usage-cli)
   - [Basic usage](#basic-usage-1)
   - [Repeated fields](#repeated-fields-1)
   - [Enum fields](#enum-fields-1)
   - [Bytes type fields](#bytes-type-fields-1)
   - [Client streaming RPC](#client-streaming-rpc-1)
   - [Server streaming RPC](#server-streaming-rpc-1)
   - [Bidirectional streaming RPC](#bidirectional-streaming-rpc-1)
   - [Enriched response](#enriched-response-1)
- [Other features](#other-features)
   - [gRPC-Web](#grpc-web)
- [Supported IDL (interface definition language)](#supported-idl-interface-definition-language)
- [Supported Codec](#supported-codec)
- [Supported Compressor](#supported-compressor)
- [See Also](#see-also)


## Installation
Highly recommended methods are GitHub Releases or Homebrew because these can be updated automatically by the built-in feature in Evans.  

### From GitHub Releases
Please see [GitHub Releases](https://github.com/ktr0731/evans/releases).  
Available binaries are:
- macOS
- Linux
- Windows

### macOS
``` sh
$ brew tap ktr0731/evans
$ brew install evans
```

### **[Not-recommended]** go get
Go v1.13 (with mod-aware mode) or later is required.  
`go get` installation is not supported officially.
``` sh
$ go get github.com/ktr0731/evans
```

## Usage (REPL)
### Basic usage
Evans consists of some commands in REPL mode.  

The proto file which read in the demonstration and its implementation are available at [ktr0731/grpc-test](https://github.com/ktr0731/grpc-test).  
`grpc-test`'s details can see `grpc-test --help`.

Enter to REPL.
``` sh
$ cd grpc-test
$ evans repl api/api.proto
```

If your server is enabling [gRPC reflection](https://github.com/grpc/grpc/blob/master/doc/server-reflection.md), you can launch Evans with only `-r` (`--reflection`) option.
``` sh
$ evans -r repl
```

Also if the server requires secure TLS connections, you can launch Evans with the `-t` (`--tls`) option.
``` sh
$ evans --tls --host example.com -r repl
```

To show package names of proto files REPL read:  
```
> show package
+---------+
| PACKAGE |
+---------+
| api     |
+---------+
```

To show the summary of services or messages:
```
> package api
> show service
+---------+----------------------+-----------------------------+----------------+
| SERVICE |         RPC          |         REQUESTTYPE         |  RESPONSETYPE  |
+---------+----------------------+-----------------------------+----------------+
| Example | Unary                | SimpleRequest               | SimpleResponse |
|         | UnaryMessage         | UnaryMessageRequest         | SimpleResponse |
|         | UnaryRepeated        | UnaryRepeatedRequest        | SimpleResponse |
|         | UnaryRepeatedMessage | UnaryRepeatedMessageRequest | SimpleResponse |
|         | UnaryRepeatedEnum    | UnaryRepeatedEnumRequest    | SimpleResponse |
|         | UnarySelf            | UnarySelfRequest            | SimpleResponse |
|         | UnaryMap             | UnaryMapRequest             | SimpleResponse |
|         | UnaryMapMessage      | UnaryMapMessageRequest      | SimpleResponse |
|         | UnaryOneof           | UnaryOneofRequest           | SimpleResponse |
|         | UnaryEnum            | UnaryEnumRequest            | SimpleResponse |
|         | UnaryBytes           | UnaryBytesRequest           | SimpleResponse |
|         | ClientStreaming      | SimpleRequest               | SimpleResponse |
|         | ServerStreaming      | SimpleRequest               | SimpleResponse |
|         | BidiStreaming        | SimpleRequest               | SimpleResponse |
+---------+----------------------+-----------------------------+----------------+

> show message
+-----------------------------+
|           MESSAGE           |
+-----------------------------+
| SimpleRequest               |
| SimpleResponse              |
| Name                        |
| UnaryMessageRequest         |
| UnaryRepeatedRequest        |
| UnaryRepeatedMessageRequest |
| UnaryRepeatedEnumRequest    |
| UnarySelfRequest            |
| Person                      |
| UnaryMapRequest             |
| UnaryMapMessageRequest      |
| UnaryOneofRequest           |
| UnaryEnumRequest            |
| UnaryBytesRequest           |
+-----------------------------+
```

To show more description of a message:  
```
> desc SimpleRequest
+-------+-------------+
| FIELD |    TYPE     |
+-------+-------------+
| name  | TYPE_STRING |
+-------+-------------+
```

Set headers for each request:
```
> header foo=bar
```

To show headers:
```
> show header
+-------------+-------+
|     KEY     |  VAL  |
+-------------+-------+
| foo         | bar   |
| grpc-client | evans |
+-------------+-------+
```

Note that if you want to set comma-included string to a header value, it is required to specify `--raw` option.

To remove the added header:
```
> header foo
> show header
+-------------+-------+
|     KEY     |  VAL  |
+-------------+-------+
| grpc-client | evans |
+-------------+-------+
```

Call a RPC:  
```
> service Example
> call Unary
name (TYPE_STRING) => ktr
{
  "message": "hello, ktr"
}

```

Evans constructs a gRPC request interactively and sends the request to a gRPC server.  
Finally, Evans prints the JSON formatted result.  

### Repeated fields
`repeated` is an array-like data structure.  
You can input some values and finish with <kbd>CTRL-D</kbd>  
```
> call UnaryRepeated
<repeated> name (TYPE_STRING) => foo
<repeated> name (TYPE_STRING) => bar
<repeated> name (TYPE_STRING) => baz
<repeated> name (TYPE_STRING) =>
{
  "message": "hello, foo, bar, baz"
}
```

### Enum fields
You can select one from the proposed selections.  
To abort it, input <kbd>CTRL-C</kbd>.
```
> call UnaryEnum
? UnaryEnumRequest  [Use arrows to move, type to filter]
> Male
  Female
{
  "message": "M"
}
```

### Bytes type fields
You can use byte literal and Unicode literal.

```
> call UnaryBytes
data (TYPE_BYTES) => \x46\x6f\x6f
{
  "message": "received: (bytes) 46 6f 6f, (string) Foo"
}

> call UnaryBytes
data (TYPE_BYTES) => \u65e5\u672c\u8a9e
{
  "message": "received: (bytes) e6 97 a5 e6 9c ac e8 aa 9e, (string) 日本語"
}
```

Or add the flag `--bytes-from-file` to read bytes from the provided relative path
```
> call UnaryBytes --bytes-from-file
data (TYPE_BYTES) => ../relative/path/to/file
```

### Client streaming RPC
Client streaming RPC accepts some requests and then returns only one response.  
Finish request inputting with <kbd>CTRL-D</kbd>

```
> call ClientStreaming
name (TYPE_STRING) => ktr
name (TYPE_STRING) => ktr
name (TYPE_STRING) => ktr
name (TYPE_STRING) =>
{
  "message": "ktr, you greet 3 times."
}
```

### Server streaming RPC
Server streaming RPC accepts only one request and then returns some responses.
Each response is represented as another JSON formatted output.
```
name (TYPE_STRING) => ktr
{
  "message": "hello ktr, I greet 0 times."
}

{
  "message": "hello ktr, I greet 1 times."
}
```

### Bidirectional streaming RPC
Bidirectional streaming RPC accepts some requests and returns some responses corresponding to each request.
Finish request inputting with <kbd>CTRL-D</kbd>

```
> call BidiStreaming
name (TYPE_STRING) => foo
{
  "message": "hello foo, I greet 0 times."
}

{
  "message": "hello foo, I greet 1 times."
}

{
  "message": "hello foo, I greet 2 times."
}

name (TYPE_STRING) => bar
{
  "message": "hello bar, I greet 0 times."
}

name (TYPE_STRING) =>
```

### Skip the rest of the fields
Evans recognizes <kbd>CTRL-C</kbd> as a special key that skips the rest of the fields in the current message type.
For example, we assume that we are inputting `Request` described in the following message:

``` proto
message FullName {
  string first_name = 1;
  string last_name = 2;
}

message Request {
  string nickname = 1;
  FullName full_name = 2;
}
```

If we enter <kbd>CTRL-C</kbd> at the following moment, `full_name` field will be skipped.

```
nickname (TYPE_STRING) =>
```

The actual request value is just like this.

``` json
{}
```

If we enter <kbd>CTRL-C</kbd> at the following moment, `last_name` field will be skipped.

```
nickname (TYPE_STRING) => myamori
full_name::first_name (TYPE_STRING) => aoi
full_name::last_name (TYPE_STRING) =>
```

The actual request value is just like this.

``` json
{
  "nickname": "myamori",
  "fullName": {
    "firstName": "aoi"
  }
}
```

By default, Evans digs down each message field automatically.  
For example, we assume that we are inputting `Request` described in the following message:

``` proto
message FullName {
  string first_name = 1;
  string last_name = 2;
}

message Request {
  FullName full_name = 1;
}
```

In this case, REPL prompts `full_name.first_name` automatically. To skip `full_name` itself, we can use `--dig-manually` option.
It asks whether dig down a message field when the prompt encountered it.

### Enriched response
To display more enriched response, you can use `--enrich` option.

```
> call --enrich Unary
name (TYPE_STRING) => ktr
content-type: application/grpc
header_key1: header_val1
header_key2: header_val2

{
  "message": "hello, ktr"
}

trailer_key1: trailer_val1
trailer_key2: trailer_val2

code: OK
number: 0
message: ""
```

## Usage (CLI)
### Basic usage
CLI mode also has some commands.  

`list` command provides gRPC service inspection against to the gRPC server.

``` sh
$ evans -r cli list
api.Example
grpc.reflection.v1alpha.ServerReflection
```

If an service name is specified, it displays methods belonging to the service.

``` sh
$ evans -r cli list api.Example
api.Example.Unary
api.Example.UnaryBytes
api.Example.UnaryEnum
...
```

`desc` command describes the passed symbol (service, method, message, and so on).

``` sh
api.Example:
service Example {
  rpc Unary ( .api.SimpleRequest ) returns ( .api.SimpleResponse );
  rpc UnaryBytes ( .api.UnaryBytesRequest ) returns ( .api.SimpleResponse );
  rpc UnaryEnum ( .api.UnaryEnumRequest ) returns ( .api.SimpleResponse );
  ...
}
```

`call` command invokes a method.
You can input requests from `stdin` or files.  

Use `--file` (`-f`) to specify a file.
``` sh
$ cat request.json
{
  "name": "ktr"
}

$ evans --proto api/api.proto cli call --file request.json api.Example.Unary
{
  "message": "hello, ktr"
}
```

If gRPC reflection is enabled, `--reflection` (`-r`) is available instead of specifying proto files.

``` sh
$ evans -r cli call --file request.json api.Example.Unary
{
  "message": "hello, ktr"
}
```

Use `stdin`.
``` sh
$ echo '{ "name": "ktr" }' | evans cli call api.Example.Unary
{
  "message": "hello, ktr"
}
```

If `.evans.toml` is exist in Git project root, you can denote default values.  

``` toml
[default]
protoFile = ["api/api.proto"]
package = "api"
service = "Example"
```

Then, the command will be more clear.  

``` sh
$ echo '{ "name": "ktr" }' | evans cli call Unary
{
  "message": "hello, ktr"
}
```

### Repeated fields
``` sh
$ echo '{ "name": ["foo", "bar"] }' | evans -r cli call api.Example.UnaryRepeated
{
  "message": "hello, foo, bar"
}
```

### Enum fields
``` sh
$ echo '{ "gender": 0 }' | evans -r cli call api.Example.UnaryEnum
{
  "message": "M"
}
```

### Bytes type fields
You need to encode bytes by Base64.  
This constraint is come from Go's standard package [encoding/json](https://golang.org/pkg/encoding/json/#Marshal)  
``` sh
$ echo 'Foo' | base64
Rm9vCg==

$ echo '{"data": "Rm9vCg=="}' | evans -r cli call api.Example.UnaryBytes
```

### Client streaming RPC
``` sh
$ echo '{ "name": "ktr" } { "name": "ktr" }' | evans -r cli call api.Example.ClientStreaming
{
  "message": "ktr, you greet 2 times."
}
```

### Server streaming RPC
``` sh
$ echo '{ "name": "ktr" }' | evans -r cli call api.Example.ServerStreaming
{
  "message": "hello ktr, I greet 0 times."
}

{
  "message": "hello ktr, I greet 1 times."
}

{
  "message": "hello ktr, I greet 2 times."
}
```

### Bidirectional streaming RPC
``` sh
$ echo '{ "name": "foo" } { "name": "bar" }' | evans -r cli call api.Example.BidiStreaming
{
  "message": "hello foo, I greet 0 times."
}

{
  "message": "hello foo, I greet 1 times."
}

{
  "message": "hello foo, I greet 2 times."
}

{
  "message": "hello foo, I greet 3 times."
}

{
  "message": "hello bar, I greet 0 times."
}
```

### Enriched response
To display more enriched response, you can use `--enrich` option.

``` 
$ echo '{"name": "ktr"}' | evans -r cli call --enrich api.Example.Unary                                                                                     ~/.ghq/src/github.com/ktr0731/grpc-test master
content-type: application/grpc
header_key1: header_val1
header_key2: header_val2

{
  "message": "hello, ktr"
}

trailer_key1: trailer_val1
trailer_key2: trailer_val2

code: OK
number: 0
message: ""
```

JSON output is also available with `--json` option.

## Other features
### gRPC-Web
Evans also support gRPC-Web protocol.  
Tested gRPC-Web implementations are:
- [improbable-eng/grpc-web](https://github.com/improbable-eng/grpc-web)

At the moment TLS is not supported for gRPC-Web.

## Supported IDL (interface definition language)
- [Protocol Buffers 3](https://developers.google.com/protocol-buffers/)  

## Supported Codec
- [Protocol Buffers 3](https://developers.google.com/protocol-buffers/)  

## Supported Compressor
- [GZIP](https://godoc.org/google.golang.org/grpc/encoding/gzip)  

## See Also
Evans (DJ YOSHITAKA)  
[![Evans](https://user-images.githubusercontent.com/12953836/47862601-da7d9c00-de38-11e8-80be-9fc981903f6c.png)](https://itunes.apple.com/jp/album/jubeat-original-soundtrack/id325295989)
