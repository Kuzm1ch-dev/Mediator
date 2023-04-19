package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mediator/src/client"
	"mediator/src/config"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var Transactions map[string]chan ([]byte) = make(map[string]chan ([]byte))

type Manager struct {
	Config config.Config
	Gin    *gin.Engine
	Http   *client.Http `json:"Http, omitempty"`
}

func InitManager() *Manager {
	pwd, _ := os.Getwd()
	configFile, _ := ioutil.ReadFile(pwd + "/config.json")
	var config config.Config
	err := json.Unmarshal(configFile, &config)
	if err != nil {
		//file, _ := os.OpenFile(pwd+"/config.json", os.O_RDWR|os.O_CREATE, 0777)
		//file.Write([]byte("{\"Host\" : \"localhost\",\"Port\" : 8081, \"Client\": { \"Bitrix\":{ \"Name\": \"TEST\", \"UUID\": \"TEST\", \"URL\": \"API URL\", \"Topics\": [ { \"Name\": \"TOPIC\", \"Methods\": [ \"Add\", \"Update\", \"Remove\" ] } ], \"Subscriptions\": [ { \"Name\": \"TOPIC2\", \"Methods\": [ { \"Name\": \"Add\", \"OperatingMethod\": \"create.json\", \"WithoutTopic\": true, \"Headers\" : { \"Content-Type\" : \"application/json\", \"Accept\" : \"application/json\" } }, { \"Name\": \"Update\", \"OperatingMethod\": \"update.json\", \"WithoutTopic\": true, \"Headers\" : { \"Content-Type\" : \"application/json\", \"Accept\" : \"application/json\" } }, { \"Name\": \"Delete\", \"OperatingMethod\": \"remove.json\", \"WithoutTopic\": true, \"Headers\" : { \"Content-Type\" : \"application/json\", \"Accept\" : \"application/json\" } } ] } ] } } }"))
		//file.Close()
		log.Println(err, "Заполните config.json файл")
	}
	router := gin.Default()

	manager := &Manager{config, router, client.InitHttp()}

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
			for _, method := range topic.Methods {
				manager.Gin.POST(fmt.Sprintf("%s/%s/%s", client.UUID, topic.Name, method.Name), manager.HandleContext)
			}
		}
	}

	return manager
}

func (m *Manager) HandleContext(ctx *gin.Context) {
	a := strings.Split(ctx.Request.URL.String(), "/")
	name, topic, method := a[1], a[2], a[3]
	source_client, err := m.Config.GetClient(name)

	data, _ := ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		log.Println("Нет такого источника")
		return
	}
	source_topic, err := source_client.GetTopic(topic)
	if err != nil {
		log.Println("Нет такого топика")
		return
	}
	source_method, err := source_topic.GetMethod(method)
	if err != nil {
		log.Println("Нет такого метода")
		return
	}

	var channel chan ([]byte) // Канал для возврата данных
	var waitResponse bool = false
	if ctx.Request.Header.Get("Response-Header") != "" { // Если от этого запроса ожидают ответа, берем канал из буфера
		channel = Transactions[ctx.Request.Header.Get("Response-Header")]
		waitResponse = true
	}
	var additionalHeaders map[string]string
	additionalHeaders = make(map[string]string)

	if source_method.Response { // Если запрос с ожиданием ответа создаем канал и кладем в буфер
		uuidTempRoute := uuid.NewString()
		additionalHeaders["Response-Header"] = uuidTempRoute
		Transactions[uuidTempRoute] = make(chan ([]byte))
		go func() {
			log.Println(source_client.Name, "Пауза. Ожидаю ответ!")
			response := <-Transactions[uuidTempRoute]
			log.Println(string(response))
			log.Println("Возобновление. Получил ответ!")
			ctx.Data(200, "json", []byte(response))
		}()
	}

	for _, destination_client := range m.Config.ClientList {
		if destination_client.Name == name {
			continue
		}
		om, withoutTopic := destination_client.HasSubscription(topic, method)
		log.Println(fmt.Sprintf("from %s to %s, topic: %s method: %s", name, destination_client.Name, topic, om))
		var resp *http.Response
		if om != "" { // Проверка на стороннее название метода REST у точки назначения
			if withoutTopic {
				resp, err = m.Http.Do(fmt.Sprintf("%s/%s", destination_client.URL, om), data, destination_client.GetHeaders(topic, method), additionalHeaders)
			} else {
				resp, err = m.Http.Do(fmt.Sprintf("%s/%s/%s", destination_client.URL, topic, om), data, destination_client.GetHeaders(topic, method), additionalHeaders)
			}
			if err != nil {
				log.Println(err)
				return
			}
			if waitResponse {
				data_response, _ := ioutil.ReadAll(resp.Body)
				channel <- data_response
				return
			}
			data, _ := ioutil.ReadAll(resp.Body)
			log.Println(source_client.Name, string(data))
			ctx.Data(200, "json", data)
		}
	}
}
func (m *Manager) Send(source_client *client.Client, message []byte, topic string, method string) {

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
