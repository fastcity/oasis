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
	SendMsg(topic, []byte) error
	Close()
}

// NewKafkaConfig
func NewKafkaModel(Addrs []string) KInterface {
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

func (k *kkModel) SendMsg(topic string, data []byte) error {
	msg := &sarama.ProducerMessage{}
	msg.Topic = topic
	msg.Value = sarama.ByteEncoder(data)

	// defer client.Close()
	pid, offset, err := k.Producer.SendMessage(msg)

	if err != nil {
		return err
	}
	fmt.Printf("分区ID:%v, offset:%v \n", pid, offset)
	return nil
}

func (k *kkModel) Close() {
	if k != nil && k.Producer != nil {
		k.Producer.Close()
	}

}
