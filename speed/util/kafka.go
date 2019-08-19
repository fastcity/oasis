package util

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
)

type kaModel struct {
	err error
	// Topic string
	addrs []string
	// Msg   string
	config   *sarama.Config
	producer sarama.SyncProducer
	consumer sarama.Consumer

	kaTopics []string // kakfa 存在的topic

	topics []string // 自己设置的监听的topic
	keys   map[string]chan []byte
}

// KaInterface KaInterface
type KaInterface interface {
	SetKeys([]string) KaInterface
	AddKey(string, chan []byte) KaInterface
	SetTopics([]string) KaInterface
	SendMsg(string, string, string) error
	ReciveMsg(chan []byte)
	GetKeyMsg(key string) chan []byte
	Close()
	Error() error
}

// NewProducer NewProducer
func NewProducer(addrs []string) KaInterface {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          //赋值为-1：这意味着producer在follower副本确认接收到数据后才算一次发送完成。
	config.Producer.Partitioner = sarama.NewRandomPartitioner //写到随机分区中，默认设置8个分区
	config.Producer.Return.Successes = true
	client, err := sarama.NewSyncProducer(addrs, config)
	if err != nil {
		logger.Error("sarama.NewSyncProducer error:", err)
		return &kaModel{err: err}
	}
	m := &kaModel{
		config:   config,
		producer: client,
		addrs:    addrs,
	}
	return m
}

// NewConsumer NewConsumer
func NewConsumer(addrs []string) KaInterface {
	consumer, err := sarama.NewConsumer(addrs, nil)

	if err != nil {
		fmt.Println("sarama.NewConsumer error:", err)
		return &kaModel{}
	}
	m := &kaModel{
		consumer: consumer,
		addrs:    addrs,
	}
	fmt.Println("m NewConsumer", m.consumer)
	return m
}

func (k *kaModel) SetChain(chain string) KaInterface {
	// k.chain = chain
	return k
}
func (k *kaModel) SetTopics(topics []string) KaInterface {
	k.topics = topics
	return k
}
func (k *kaModel) SetKeys(keys []string) KaInterface {
	if k.keys == nil {
		k.keys = make(map[string]chan []byte, len(keys))
	}
	for _, key := range keys {
		ch := make(chan []byte)
		k.keys[key] = ch
		return k
	}

	return k
}

func (k *kaModel) AddKey(key string, ch chan []byte) KaInterface {
	if k.keys == nil {
		k.keys = make(map[string]chan []byte)
	}
	k.keys[key] = ch
	return k
}

func (k *kaModel) SendMsg(key, topic, data string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(data),
	}

	// defer client.Close()
	pid, offset, err := k.producer.SendMessage(msg)

	if err != nil {
		return err
	}
	fmt.Printf("分区ID:%v, offset:%v \n", pid, offset)
	return nil
}

func (k *kaModel) ReciveMsg(msgValue chan []byte) {
	if k.consumer == nil {
		fmt.Println("sarama.NewConsumer  k.Consumer==nil:")
	}
	kaTopics, err := k.consumer.Topics()
	if err != nil {
		fmt.Println("get topics error")
	}

	k.kaTopics = kaTopics

	// msgKey := make(chan []byte)

	for _, topic := range k.topics {
		fmt.Println("topic", topic)

		if !k.topicExist(topic) {
			fmt.Println("topic is not exist on kafka", topic)
			continue
		}
		//  一个 topic 一个 协程

		partitionList, err := k.consumer.Partitions(topic)

		if err != nil {
			fmt.Println(" get topic Partitions error:", err)
		}
		fmt.Println("partitionList:", partitionList)
		for partition := range partitionList {
			pc, err := k.consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
			if err != nil {
				fmt.Println("sarama ConsumePartition error:", err)
			}

			defer pc.AsyncClose()

			for {
				select {
				case msg := <-pc.Messages():

					k.keys[string(msg.Key)] <- msg.Value
					// switch string(msg.Key) {
					// case "TX":
					// 	msgValue <- msg.Value

					// }

					fmt.Printf("msg offset: %d, partition: %d, timestamp: %s,key:%s, value: %s\n", msg.Offset, msg.Partition, msg.Timestamp.String(), string(msg.Key), string(msg.Value))
				case err := <-pc.Errors():
					fmt.Printf("err :%s\n", err.Error())
				}
			}
		}
	}
	// return msgKey
}

func (k *kaModel) GetKeyMsg(key string) chan []byte {
	return k.keys[key]
}

func (k *kaModel) Close() {
	if k != nil {
		if k.consumer != nil {
			k.consumer.Close()
		}
		if k.consumer != nil {
			k.consumer.Close()
		}
	}
	// k.Consumer.Close()
}

func (k *kaModel) Error() error {
	if k == nil {
		return nil
	}

	return k.err
}

func findAll(strs []string) []string {
	s := []string{}
	for _, str := range strs {
		if strings.HasPrefix(str, "VCT") {
			s = append(s, str)
		}
	}
	return s
}

func (k *kaModel) topicExist(topic string) bool {

	for _, top := range k.kaTopics {
		if top == topic {
			return true
		}
	}
	return false
}
