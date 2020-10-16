# amqping
Test utility for AMQP proto

Requirements: go, [amqp](github.com/streadway/amqp)
Build: `go build -o amqping main.go`

```
Usage of ./amqping:
  -act string
    	send/recieve/both (default "both")
  -auth string
    	login:pass (default "guest:guest")
  -hst string
    	host:port (default "localhost:5672")
  -inf
    	send msg infinitely
  -msg string
    	message to push to queue (default "knock_knock")
  -q string
    	name of queue to push to (default "default")
  -tls
    	port is TLS
```

To use `amqps://`, one needs to place cert, ca and key to `ssl` folder with corresponding naming.
