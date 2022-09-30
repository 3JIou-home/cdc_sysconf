package main

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"regexp"
	"time"
)

type MysqlCDR struct {
	After struct {
		Accountcode   string    `json:"accountcode"`
		Amaflags      int       `json:"amaflags"`
		Billsec       int       `json:"billsec"`
		Calldate      time.Time `json:"calldate"`
		Channel       string    `json:"channel"`
		Clid          string    `json:"clid"`
		Cnam          string    `json:"cnam"`
		Cnum          string    `json:"cnum"`
		Dcontext      string    `json:"dcontext"`
		Did           string    `json:"did"`
		Disposition   string    `json:"disposition"`
		Dst           string    `json:"dst"`
		DstCnam       string    `json:"dst_cnam"`
		Dstchannel    string    `json:"dstchannel"`
		Duration      int       `json:"duration"`
		Lastapp       string    `json:"lastapp"`
		Lastdata      string    `json:"lastdata"`
		OutboundCnam  string    `json:"outbound_cnam"`
		OutboundCnum  string    `json:"outbound_cnum"`
		Recordingfile string    `json:"recordingfile"`
		Src           string    `json:"src"`
		Uniqueid      string    `json:"uniqueid"`
		Userfield     string    `json:"userfield"`
	} `json:"after"`
	Before interface{} `json:"before"`
	Op     string      `json:"op"`
	Source struct {
		Connector string      `json:"connector"`
		Db        string      `json:"db"`
		File      string      `json:"file"`
		Gtid      interface{} `json:"gtid"`
		Name      string      `json:"name"`
		Pos       int         `json:"pos"`
		Query     interface{} `json:"query"`
		Row       int         `json:"row"`
		Sequence  interface{} `json:"sequence"`
		ServerID  int         `json:"server_id"`
		Snapshot  string      `json:"snapshot"`
		Table     string      `json:"table"`
		Thread    interface{} `json:"thread"`
		TsMs      int64       `json:"ts_ms"`
		Version   string      `json:"version"`
	} `json:"source"`
	Transaction interface{} `json:"transaction"`
	TsMs        int64       `json:"ts_ms"`
}

var (
	duration = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "duration",
		Help: "duration by dstchannel counter",
	}, []string{"provider", "dst_provider"})
	callFail = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "call_fail",
		Help: "failed call counter",
	}, []string{"provider", "dst_provider"})
)

func CreateRedPandaConsumer() *kafka.Consumer {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "host-of-redpanda|kafka:9092",
		"group.id":          "mysql-cdr",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	return consumer
}

func CheckConnector() {
	response, err := http.Get("http://host-of-debezium/connectors/mysql-cdr")
	defer response.Body.Close()

	if err != nil {
		panic(err)
	}
	if response.StatusCode != 200 {
		log.Fatalln("Connector not exist")
		// TODO: Add create connector if not exist
		//RegisterConnector()
	}
}

func SubscribeTopic(consumer *kafka.Consumer) {
	consumer.Subscribe("mysql-cdr.my-shit-queue", nil)

	fmt.Println("Subscribed to product topic")
}

func CloseConsumer(consumer *kafka.Consumer) {
	consumer.Close()
}

func DistProvider(parser []string) string {
	if parser[1] != "" {
		return "kcell"
	} else if parser[2] != "" {
		return "mobile"
	}
	return "city"
}

func ReadTopicMessagesAndAddMetrics(consumer *kafka.Consumer) {
	type tmpJsonStruct struct {
		MysqlCDR MysqlCDR `json:"payload"`
	}
	var payloadInterface tmpJsonStruct
	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			json.Unmarshal(msg.Value, &payloadInterface)
			reProv := regexp.MustCompile(".*\\/(.*)\\-")
			reNumber := regexp.MustCompile("\\.*(701|775|778)|(70[0-9]|747|75[0-1]|76[0-4]|77[2-4, 6-7,9]|771|777)[[:digit:]]{7}$")
			var lablesParsed []string
			matchProvider := reProv.FindStringSubmatch(payloadInterface.MysqlCDR.After.Dstchannel)
			matchNumber := reNumber.FindStringSubmatch(payloadInterface.MysqlCDR.After.Dst)
			if payloadInterface.MysqlCDR.Op == "c" &&
				matchProvider != nil && matchNumber != nil {
				lablesParsed = append(lablesParsed, matchProvider[1])
				lablesParsed = append(lablesParsed, DistProvider(matchNumber))
				duration.WithLabelValues(lablesParsed...).Add(float64(payloadInterface.MysqlCDR.After.Billsec))
				if payloadInterface.MysqlCDR.After.Disposition == "FALIED" {
					callFail.WithLabelValues(lablesParsed...).Add(float64(payloadInterface.MysqlCDR.After.Billsec))
				}
			}
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
	CloseConsumer(consumer)
}

func main() {
	consumer := CreateRedPandaConsumer()
	CheckConnector()
	SubscribeTopic(consumer)
	go ReadTopicMessagesAndAddMetrics(consumer)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
	defer CloseConsumer(consumer)
}
