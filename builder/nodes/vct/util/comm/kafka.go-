package comm

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
)

type kkModel struct {
	// Topic string
	Addrs []string
	// Msg   string
	Config   *sarama.Config
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	topics   []string
	keys     map[string]chan []byte
}

type KInterface interface {
	SetKeys([]string) KInterface
	AddKey(string, chan []byte) KInterface
	SetTopics([]string) KInterface
	SendMsg(string, string) error
	ReciveMsg(chan []byte)
	Close()
}

// NewProducer NewProducer
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

// NewConsumer NewConsumer
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
	fmt.Println("m NewConsumer", m.Consumer)
	return m
}

func (k *kkModel) SetChain(chain string) KInterface {
	// k.chain = chain
	return k
}
func (k *kkModel) SetTopics(topics []string) KInterface {
	k.topics = topics
	return k
}
func (k *kkModel) SetKeys(keys []string) KInterface {
	// k.keys = keys
	return k
}

func (k *kkModel) AddKey(key string, ch chan []byte) KInterface {
	if k.keys == nil {
		k.keys = make(map[string]chan []byte)
	}
	k.keys[key] = ch
	return k
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

func (k *kkModel) ReciveMsg(msgValue chan []byte) {
	if k.Consumer == nil {
		fmt.Println("sarama.NewConsumer  k.Consumer==nil:")
	}
	// ts, err := k.Consumer.Topics()
	// if err != nil {
	// 	fmt.Println("get topics error")
	// }
	// tss := findAll(ts)
	// fmt.Println("topics", ts, tss)
	// msgKey := make(chan []byte)

	for _, topic := range k.topics {
		fmt.Println("topic", topic)
		partitionList, err := k.Consumer.Partitions(topic)

		if err != nil {
			fmt.Println("sarama.NewConsumer error:", err)
		}
		fmt.Println("partitionList:", partitionList)
		for partition := range partitionList {
			pc, err := k.Consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
			if err != nil {
				fmt.Println("sarama.NewConsumer error:", err)
				// panic(err)
			}

			defer pc.AsyncClose()

			// wg.Add(1)
			for {
				select {
				case msg := <-pc.Messages():

					k.keys[string(msg.Key)] <- msg.Value
					// switch string(msg.Key) {
					// case "TX":
					// 	msgValue <- msg.Value

					// }

					fmt.Printf("msg offset: %d, partition: %d, timestamp: %s, value: %s\n", msg.Offset, msg.Partition, msg.Timestamp.String(), string(msg.Value))
				case err := <-pc.Errors():
					fmt.Printf("err :%s\n", err.Error())
				}
			}
		}
	}
	// return msgKey
}

func (k *kkModel) Close() {
	if k != nil {

		if k.Producer != nil {
			k.Producer.Close()
		}
		if k.Consumer != nil {
			k.Consumer.Close()
		}
	}

	// k.Consumer.Close()
}

func findAll(strs []string) []string {
	s := []string{}
	for _, str := range strs {
		if strings.HasPrefix(str, "VCT") {
			s = append(s, str)
		}
		// if str == "VCT_TX" {
		// 	s = append(s, str)
		// }
	}
	return s

}
