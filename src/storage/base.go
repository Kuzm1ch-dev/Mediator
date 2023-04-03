package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Base struct {
	OffsetStorage Topics
	MessageStroge Messages
}

type Topics struct {
	Storage map[string]Topic
}

type Messages struct {
	Index   int
	Storage map[int]Message
}

type Message struct {
	Topic string
	Data  string
}

type Topic struct {
	Storage map[string]int
}

func InitBase() *Base {
	base := &Base{}
	base.OffsetStorage = Topics{make(map[string]Topic)}
	base.MessageStroge = Messages{0, make(map[int]Message)}
	base.Load()
	return base
}

func (b *Base) AddTopic(topic string) {
	b.OffsetStorage.Storage[topic] = Topic{make(map[string]int)}
	log.Println(fmt.Sprintf("Топик %s создан в базе", topic))
	b.Save()
}

func (b *Base) AddSubscription(topic string, sub string) {
	var topicMap = b.OffsetStorage.Storage[topic]
	topicMap.Storage[sub] = 0
	log.Println(fmt.Sprintf("Подписчик %s на топик %s создан в базе", sub, topic))
	b.Save()
}

func (b *Base) AddOffset(topic string, sub string) {
	var topicMap = b.OffsetStorage.Storage[topic]
	topicMap.Storage[sub] += 1
	log.Println(sub, " отстал на ", topicMap.Storage[sub])
	b.Save()
}

func (b *Base) SubOffset(topic string, sub string) {
	var topicMap = b.OffsetStorage.Storage[topic]
	topicMap.Storage[sub] -= 1
	b.Save()
}

func (b *Base) AddMessage(message string, topic string) {
	var index = b.MessageStroge.Index + 1
	b.MessageStroge.Storage[index] = Message{topic, message}
	b.Save()
}

func (b *Base) Save() {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	enc.Encode(b)
	pwd, _ := os.Getwd()
	file, _ := os.OpenFile(pwd+"/data.bin", os.O_RDWR|os.O_CREATE, 0777)
	file.Write(buff.Bytes())
	file.Close()
}

func (b *Base) Load() {
	pwd, _ := os.Getwd()
	databin, err := ioutil.ReadFile(pwd + "/data.bin")
	if err != nil {
		log.Println(err)
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(databin))
	dec.Decode(b)
}
