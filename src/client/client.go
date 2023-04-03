package client

type Config struct {
	Host       string
	Port       int
	ClientList map[string]*Client `json:"Client"`
}

type Method struct {
	Name            string            `json:"Name"`
	OperatingMethod string            `json:"OperatingMethod,omitempty"`
	WithoutTopic    bool              `json:"WithoutTopic,omitempty"`
	Headers         map[string]string `json:"Headers,omitempty"`
}

type Topic struct {
	Name    string   `json:"Name"`
	Methods []string `json:"Methods"`
}

type Subscription struct {
	Name    string   `json:"Name"`
	Methods []Method `json:"Methods"`
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
