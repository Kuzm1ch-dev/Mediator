package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mediator/src/client"
	"mediator/src/storage"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Manager struct {
	Config client.Config
	Gin    *gin.Engine
	Http   *client.Http `json:"Http, omitempty"`
	Base   storage.Base
}

func InitManager() *Manager {
	pwd, _ := os.Getwd()
	configFile, _ := ioutil.ReadFile(pwd + "/config.json")
	var config client.Config
	err := json.Unmarshal(configFile, &config)
	if err != nil {
		file, _ := os.OpenFile(pwd+"/config.json", os.O_RDWR|os.O_CREATE, 0777)
		file.Write([]byte("{\"Host\" : \"localhost\",\"Port\" : 8081, \"Client\": { \"Bitrix\":{ \"Name\": \"TEST\", \"UUID\": \"TEST\", \"URL\": \"API URL\", \"Topics\": [ { \"Name\": \"TOPIC\", \"Methods\": [ \"Add\", \"Update\", \"Remove\" ] } ], \"Subscriptions\": [ { \"Name\": \"TOPIC2\", \"Methods\": [ { \"Name\": \"Add\", \"OperatingMethod\": \"create.json\", \"WithoutTopic\": true, \"Headers\" : { \"Content-Type\" : \"application/json\", \"Accept\" : \"application/json\" } }, { \"Name\": \"Update\", \"OperatingMethod\": \"update.json\", \"WithoutTopic\": true, \"Headers\" : { \"Content-Type\" : \"application/json\", \"Accept\" : \"application/json\" } }, { \"Name\": \"Delete\", \"OperatingMethod\": \"remove.json\", \"WithoutTopic\": true, \"Headers\" : { \"Content-Type\" : \"application/json\", \"Accept\" : \"application/json\" } } ] } ] } } }"))
		file.Close()
		log.Println("Заполните config.json файл")
	}
	router := gin.Default()

	manager := &Manager{config, router, client.InitHttp(), *storage.InitBase()}

	for _, client := range manager.Config.ClientList {
		log.Println("Служебный топик для клиента:")
		manager.Gin.POST(fmt.Sprintf("%s/%s", client.UUID, "Query"), manager.HandleQuery)

		/*
			"Топик" Query используется в том случае, если полученных данных недостаточно
			и необходимо сделать запрос клиенту для уточнения.

		*/

		for _, topic := range client.Topics {
			if topic.Name == "Query" {
				return nil
			}
			manager.Base.AddTopic(topic.Name)
			for _, method := range topic.Methods {
				manager.Gin.POST(fmt.Sprintf("%s/%s/%s", client.UUID, topic.Name, method), manager.HandleContext)
			}
		}
	}
	for _, client := range manager.Config.ClientList {
		for _, subscription := range client.Subscriptions {
			manager.Base.AddSubscription(subscription.Name, client.Name)
		}
	}

	return manager
}

func (m *Manager) HandleContext(ctx *gin.Context) {
	a := strings.Split(ctx.Request.URL.String(), "/")
	name, topic, method := a[1], a[2], a[3]

	for _, destination_client := range m.Config.ClientList {
		if destination_client.Name == name {
			continue
		}
		om, withoutTopic := destination_client.HasSubscription(topic, method)
		if om != "" {
			data, _ := ioutil.ReadAll(ctx.Request.Body)
			log.Println(fmt.Sprintf("from %s to %s, topic: %s method: %s", name, destination_client.Name, topic, om))
			//log.Println("With data: ", string(data))
			if withoutTopic {
				_, err := m.Http.Do(fmt.Sprintf("%s/%s", destination_client.URL, om), data, destination_client.GetHeaders(topic, method))
				if err != nil {
					log.Println(err)
					m.Base.AddOffset(topic, destination_client.Name)
				}
			} else {
				_, err := m.Http.Do(fmt.Sprintf("%s/%s/%s", destination_client.URL, topic, om), data, destination_client.GetHeaders(topic, method))
				if err != nil {
					log.Println(err)
					m.Base.AddOffset(topic, destination_client.Name)
				}

			}
		}
	}
	data, _ := ioutil.ReadAll(ctx.Request.Body)
	m.Base.AddMessage(string(data), topic)
}

func (m *Manager) HandleQuery(context *gin.Context) {
	clientName := context.Request.Header.Get("Client")
	method := context.Request.Header.Get("Method")
	data, _ := ioutil.ReadAll(context.Request.Body)
	log.Println("Reqest for ", clientName, " method: ", method)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", m.Config.ClientList[clientName].URL, method), bytes.NewBuffer(data))
	if err != nil {
		log.Println(err)
	}
	req.Header.Add("Content-Type", context.Request.Header.Get("Content-Type"))
	req.Header.Add("Accept", context.Request.Header.Get("Accept"))
	resp, err := m.Http.Client.Do(req)
	jsonData, _ := ioutil.ReadAll(resp.Body)
	context.Data(200, "application/json", jsonData)
}
