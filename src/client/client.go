package client

import "errors"

type Config struct {
	Host       string
	Port       int
	ClientList map[string]*Client `json:"Client"`
}

type SMethod struct {
	Name            string            `json:"Name"`
	OperatingMethod string            `json:"OperatingMethod,omitempty"`
	WithoutTopic    bool              `json:"WithoutTopic,omitempty"`
	Headers         map[string]string `json:"Headers,omitempty"`
}

type TMethod struct {
	Name     string `json:"Name"`
	Response bool   `json:"Response,omitempty"`
}

type Topic struct {
	Name    string    `json:"Name"`
	Methods []TMethod `json:"Methods"`
}

type Subscription struct {
	Name    string    `json:"Name"`
	Methods []SMethod `json:"Methods"`
}

type Client struct {
	Name          string         `json:"Name"`
	UUID          string         `json:"UUID"`
	URL           string         `json:"URL"`
	Topics        []Topic        `json:"Topics"`
	Subscriptions []Subscription `json:"Subscriptions"`
}

func (c *Client) HasSubscription(sub string, met string) (string, bool) {
	for _, subscription := range c.Subscriptions {
		for _, method := range subscription.Methods {
			if subscription.Name == sub && method.Name == met {
				b := method.WithoutTopic
				if method.OperatingMethod == "" {
					return method.Name, b
				} else {
					return method.OperatingMethod, b
				}
			}
		}
	}
	return "", false
}

func (c *Client) GetHeaders(sub string, met string) map[string]string {
	for _, subscription := range c.Subscriptions {
		for _, method := range subscription.Methods {
			if subscription.Name == sub && method.Name == met {
				return method.Headers
			}
		}
	}
	return nil
}

func (c *Client) GetTopic(topicName string) (Topic, error) {
	for _, topic := range c.Topics {
		if topic.Name == topicName {
			return topic, nil
		}
	}
	return Topic{}, errors.New("not yet implemented")
}

func (c *Client) GetSubscription(topicName string) (Subscription, error) {
	for _, subscription := range c.Subscriptions {
		if subscription.Name == topicName {
			return subscription, nil
		}
	}
	return Subscription{}, errors.New("not yet implemented")
}

func (c *Config) GetClient(name string) (Client, error) {
	for _, client := range c.ClientList {
		if client.Name == name {
			return *client, nil
		}
	}
	return Client{}, errors.New("not yet implemented")
}

func (t *Topic) GetMethod(methodName string) (TMethod, error) {
	for _, method := range t.Methods {
		if method.Name == methodName {
			return method, nil
		}
	}
	return TMethod{}, errors.New("not yet implemented")
}

func (s *Subscription) GetMethod(methodName string) (SMethod, error) {
	for _, method := range s.Methods {
		if method.Name == methodName {
			return method, nil
		}
	}
	return SMethod{}, errors.New("not yet implemented")
}
