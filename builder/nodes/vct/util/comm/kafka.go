package comm

import (
	"fmt"

	"github.com/Shopify/sarama"
)

type kkModel struct {
	// Topic string
	Addrs []string
	// Msg   string
	Config   *sarama.Config
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
}

type KInterface interface {
	SendMsg(topic, data string) error
	Close()
}

// NewKafkaConfig
func NewProducer(Addrs []string) KInterface {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          //赋值为-1：这意味着producer在follower副本确认接收到数据后才算一次发送完成。
	config.Producer.Partitioner = sarama.NewRandomPartitioner //写到随机分区中，默认设置8个分区
	config.Producer.Return.Successes = true
	client, err := sarama.NewSyncProducer(Addrs, config)
	if err != nil {
		fmt.Println("sarama.NewSyncProducer error:", err)
		return &kkModel{}
	}
	m := &kkModel{
		Config:   config,
		Producer: client,
		Addrs:    Addrs,
	}
	return m
}

func NewConsumer(Addrs []string) KInterface {
	consumer, err := sarama.NewConsumer(Addrs, nil)

	if err != nil {
		fmt.Println("sarama.NewConsumer error:", err)
		return &kkModel{}
	}
	m := &kkModel{
		Consumer: consumer,
		Addrs:    Addrs,
	}
	return m
}

func (k *kkModel) SendMsg(topic, data string) error {
	msg := &sarama.ProducerMessage{}
	msg.Topic = topic
	msg.Value = sarama.StringEncoder(data)

	// defer client.Close()
	pid, offset, err := k.Producer.SendMessage(msg)

	if err != nil {
		return err
	}
	fmt.Printf("分区ID:%v, offset:%v \n", pid, offset)
	return nil
}

func (k *kkModel) ReciveMsg() chan []byte {
	ts, err := k.Consumer.Topics()
	if err != nil {
		fmt.Println("get topics error")
	}
	msgKey := make(chan []byte)
	for _, to := range ts {
		fmt.Println("topic", to)
		partitionList, err := k.Consumer.Partitions(to)

		if err != nil {
			fmt.Println("sarama.NewConsumer error:", err)
		}

		for partition := range partitionList {
			pc, err := k.Consumer.ConsumePartition(to, int32(partition), sarama.OffsetNewest)
			if err != nil {
				fmt.Println("sarama.NewConsumer error:", err)
				// panic(err)
			}

			defer pc.AsyncClose()

			// wg.Add(1)

			for msg := range pc.Messages() {
				msgKey <- msg.Key
				fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			}

			// go func(sarama.PartitionConsumer) {
			// 	defer wg.Done()
			// 	for msg := range pc.Messages() {
			// 		fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			// 	}
			// }(pc)
			// wg.Wait()
			// consumer.Close()
		}
	}
	return msgKey
}

func (k *kkModel) Close() {
	k.Producer.Close()
	k.Consumer.Close()
}
