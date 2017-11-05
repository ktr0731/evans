# Evans
[![wercker status](https://app.wercker.com/status/1b1e3a40c102c07ad4f61630fea6c7bf/s/master "wercker status")](https://app.wercker.com/project/byKey/1b1e3a40c102c07ad4f61630fea6c7bf)  
more expressive universal gRPC client  

## Usage
Evans consistents of some commands in REPL.  

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
