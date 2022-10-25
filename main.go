package main

import (
  "log"
  "os"
  "time"
  "math/rand"
  "flag"
  "sync"

  "crypto/x509"
  "crypto/tls"
  "io/ioutil"

  "github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
    os.Exit(1)
  }
}

func genRandStr() string {
  charset := "abcdefghijklmnopqrstuvwxyz"

  var seededRand *rand.Rand = rand.New(
    rand.NewSource(time.Now().UnixNano()))

  b := make([]byte, 10)
  for i := range b {
    b[i] = charset[seededRand.Intn(len(charset))]
  }

  return string(b)
}

func declareQueue(ch *amqp.Channel, queueArg *string) amqp.Queue {
  queue := *queueArg
  q, err := ch.QueueDeclare(
    queue, // name
    false,   // durable
    false,   // delete when unused
    false,   // exclusive
    false,   // no-wait
    nil,     // arguments
  )
  failOnError(err, "Failed to declare a queue")

  return q
}

func sendToQueue(ch *amqp.Channel, q amqp.Queue, msgArg *string, infArg *bool) {
  body := *msgArg
  infinite := *infArg

  if body == "rand" {
   body = genRandStr()
  }

  var wg sync.WaitGroup
  wg.Add(1)

  go func() {
    defer wg.Done()
    for {
      err := ch.Publish(
        "",     // exchange
        q.Name, // routing key
        false,  // mandatory
        false,  // immediate
        amqp.Publishing {
          ContentType: "text/plain",
          Body:        []byte(body),
        })
      failOnError(err, "Failed to publish a message")

      if infinite != true {
        break
      } else {
        log.Printf("Sent a message: %s", body)
        time.Sleep(1000 * time.Millisecond)
      }
    }
  }()

  wg.Wait()
}

func receiveFromQueue(ch *amqp.Channel, q amqp.Queue) {
  msgs, err := ch.Consume(
    q.Name, // queue
    "",     // consumer
    true,   // auto-ack
    false,  // exclusive
    false,  // no-local
    false,  // no-wait
    nil,    // args
  )
  failOnError(err, "Failed to register a consumer")

  forever := make(chan bool)

  go func() {
    for d := range msgs {
      log.Printf("Received a message: %s", d.Body)
    }
  }()

  log.Printf(" [*] Waiting for messages on %s. To exit press CTRL+C", q.Name)

  <-forever
}


func main() {
  actionArg := flag.String("act", "both", "send/recieve/both")
  hostArg := flag.String("hst", "localhost:5672", "host:port")
  authArg := flag.String("auth", "guest:guest", "login:pass")
  queueArg := flag.String("q", "default", "name of queue to push to")
  msgArg := flag.String("msg", "knock_knock", "message to push to queue")
  infArg := flag.Bool("inf", false, "send msg infinitely")
  tlsArg := flag.Bool("tls", false, "port is TLS")

  flag.Parse()

  var conn *amqp.Connection
  var err error

  if *tlsArg {
    //cfg := new(tls.Config)
    cfg := &tls.Config{InsecureSkipVerify: true}

    cfg.RootCAs = x509.NewCertPool()

    if ca, err := ioutil.ReadFile("ssl/ca.pem"); err == nil {
        cfg.RootCAs.AppendCertsFromPEM(ca)
    }

    if cert, err := tls.LoadX509KeyPair("ssl/cert.pem", "ssl/key.pem"); err == nil {
        cfg.Certificates = append(cfg.Certificates, cert)
    }

    conn, err = amqp.DialTLS("amqps://" + *authArg + "@" + *hostArg, cfg)
    failOnError(err, "Failed to connect over TLS to RabbitMQ")
  } else {
    conn, err = amqp.Dial("amqp://" + *authArg + "@" + *hostArg)
    failOnError(err, "Failed to connect to RabbitMQ")
  }

  defer conn.Close()

  ch, err := conn.Channel()
  failOnError(err, "Failed to open a channel")
  defer ch.Close()

  queue := declareQueue(ch, queueArg)

  if *actionArg == "both" {
    sendToQueue(ch, queue, msgArg, infArg)
    receiveFromQueue(ch, queue)
  } else if *actionArg == "receive" {
    receiveFromQueue(ch, queue)
  } else if *actionArg == "send" {
    sendToQueue(ch, queue, msgArg, infArg)
  }
}
