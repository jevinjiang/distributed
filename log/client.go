package log

import (
	"bytes"
	"distributed/registry"
	"fmt"
	stlog "log"
	"net/http"
)

func SetClientLogger(serviceURL string, clientService registry.ServiceName) {
	stlog.SetPrefix(fmt.Sprintf("[%v] - ", clientService))
	stlog.SetFlags(0)
	stlog.SetOutput(&clientLogger{url: serviceURL})
}

type clientLogger struct {
	url string
}

func (c clientLogger) Write(data []byte) (n int, err error) {
	buffer := bytes.NewBuffer([]byte(data))
	resp, err := http.Post(c.url+"/log", "text/plain", buffer)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Failed to send log message. service responded write")
	}
	return len(data), nil
}
