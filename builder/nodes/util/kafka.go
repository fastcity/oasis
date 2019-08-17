package util

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

func InitKa() {
	// conn, _ := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)

	// r := kafka.NewReader(kafka.ReaderConfig{
	// 	Brokers: []string{"localhost:9092"},
	// 	// GroupID:  "consumer-group-id",
	// 	Topic:    "VCT_TX",
	// 	MinBytes: 10e3, // 10KB
	// 	MaxBytes: 10e6, // 10MB
	// })

	// go func() {
	// 	for {
	// 		m, err := r.ReadMessage(context.Background())
	// 		if err != nil {
	// 			break
	// 		}
	// 		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	// 	}
	// }()

	// r.Close()

	// conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "VCT_TX", 0)

	conn, err := kafka.DialContext(context.Background(), "tcp", "localhost:9092")
	if err != nil {

		fmt.Println("DialLeader error", err)
	}

	partitionList, err := conn.ReadPartitions("VCT_TX")

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	for _, partition := range partitionList {
		fmt.Printf("topic:%s,pid:%d \n", partition.Topic, partition.ID)
		conn = kafka.NewConn(conn, "VCT_TX", partition.ID)
		batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max
		b := make([]byte, 10e3)            // 10KB max per message
		for {
			_, err := batch.Read(b)
			if err != nil {
				fmt.Println("error", err)
				break
			}
			fmt.Println(string(b))
		}
		batch.Close()
	}

	conn.Close()
}
