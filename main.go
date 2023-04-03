package main

import (
	"fmt"
	"mediator/src/manager"
)

func main() {
	m := manager.InitManager()
	m.Gin.Run(fmt.Sprintf("%s:%d", m.Config.Host, m.Config.Port))
}
