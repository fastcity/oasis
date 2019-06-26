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
	}
	return m
}

func NewConsumer(Addrs []string) KInterface {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          //赋值为-1：这意味着producer在follower副本确认接收到数据后才算一次发送完成。
	config.Producer.Partitioner = sarama.NewRandomPartitioner //写到随机分区中，默认设置8个分区
	config.Producer.Return.Successes = true
	client, err := sarama.NewConsumer(Addrs, config)
	if err != nil {
		fmt.Println("sarama.NewSyncProducer error:", err)
		return &kkModel{}
	}
	m := &kkModel{
		Config:   config,
		Producer: client,
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

func (k *kkModel) ReciveMsg() {
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)

	if err != nil {
		panic(err)
	}

	partitionList, err := consumer.Partitions("testGo")

	if err != nil {
		panic(err)
	}

	for partition := range partitionList {
		pc, err := consumer.ConsumePartition("testGo", int32(partition), sarama.OffsetNewest)
		if err != nil {
			panic(err)
		}

		defer pc.AsyncClose()

		wg.Add(1)

		go func(sarama.PartitionConsumer) {
			defer wg.Done()
			for msg := range pc.Messages() {
				fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			}
		}(pc)
		wg.Wait()
		consumer.Close()
	}
}

func (k *kkModel) Close() {
	k.Producer.Close()
}
