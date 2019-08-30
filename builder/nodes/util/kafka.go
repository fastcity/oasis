package util

import (
	"fmt"
	"sync"

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

	msgs chan map[string][]byte

	wg sync.WaitGroup
}

// KaInterface KaInterface
type KaInterface interface {
	SetKeys([]string) KaInterface
	AddKey(string, chan []byte) KaInterface
	SetTopics([]string) KaInterface
	SendMsg(string, string, string) error
	ReciveMsg()
	GetKeyMsg(key string) chan []byte
	GetMsg() map[string][]byte
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
		fmt.Println("sarama.NewSyncProducer error:", err)
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

func (k *kaModel) GetMsg() map[string][]byte {

	if k.msgs == nil {
		k.msgs = make(chan map[string][]byte) // 不加这个先getmsg 就卡住了，接受不了消息
	}
	select {
	case m := <-k.msgs:
		return m
	}
}

func (k *kaModel) setMsg(key string, value []byte) {
	if k.msgs == nil {
		k.msgs = make(chan map[string][]byte)
	}
	k.msgs <- map[string][]byte{
		key: value,
	}
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
	logger.Debugf("分区ID:%v, offset:%v \n", pid, offset)
	return nil
}

func (k *kaModel) ReciveMsg() {
	if k.consumer == nil {
		logger.Error("sarama.NewConsumer  k.Consumer==nil:")
	}
	kaTopics, err := k.consumer.Topics()
	if err != nil {
		logger.Error("get topics error")
	}

	k.kaTopics = kaTopics

	for _, topic := range k.topics {
		logger.Debug("topic:", topic)

		if !k.topicExist(topic) {
			logger.Error("topic is not exist on kafka", topic)
			continue
		}
		k.wg.Add(1)
		//  一个 topic 一个 协程
		go k.reciveMsgByTopic(topic)
	}

	k.wg.Wait()

}

func (k *kaModel) reciveMsgByTopic(topic string) {

	defer k.wg.Done()
	partitionList, err := k.consumer.Partitions(topic)

	if err != nil {
		logger.Error(" get topic Partitions error:", err)
	}
	logger.Debug("partitionList:", partitionList)

	for partition := range partitionList {
		pc, err := k.consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			logger.Error("sarama ConsumePartition error:", err)
		}
		// go func(pc sarama.PartitionConsumer) {
		defer pc.AsyncClose()
		for {
			select {
			case msg := <-pc.Messages():
				// k.keys[string(msg.Key)] <- msg.Value

				k.setMsg(string(msg.Key), msg.Value)

				logger.Debugf("----msg offset: %d, partition: %d, timestamp: %s,key:%s, value: %s", msg.Offset, msg.Partition, msg.Timestamp.String(), string(msg.Key), string(msg.Value))
			case err := <-pc.Errors():
				logger.Errorf("err :%s\n", err.Error())
			}
		}
		// }(pc)
	}
}

func (k *kaModel) GetKeyMsg(key string) chan []byte {
	return k.keys[key]
}

func (k *kaModel) Close() {
	if k != nil {
		if k.producer != nil {
			k.producer.Close()
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

func (k *kaModel) topicExist(topic string) bool {

	for _, top := range k.kaTopics {
		if top == topic {
			return true
		}
	}
	return false
}
