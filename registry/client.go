package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

type serviceUpdateHandler struct {
}

func (s serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dec := json.NewDecoder(r.Body)
	var p patch
	err := dec.Decode(&p)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	fmt.Printf("Update received %v \n", p)
	prov.Update(p)
}

type heartbeatHandler struct{}

func DoRegistry(r Registration) error {
	heartbeatUrl, err3 := url.Parse(r.HeartbeatUrl)
	if err3 != nil {
		return err3
	}
	serviceUpdateUrl, err2 := url.Parse(r.ServiceUpdateURL)
	if err2 != nil {
		return err2
	}
	http.HandleFunc(heartbeatUrl.Path, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	http.Handle(serviceUpdateUrl.Path, &serviceUpdateHandler{})
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(r)
	if err != nil {
		return err
	}
	resp, err := http.Post(ServiceUrl, "application/json", buf)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to registry service. Registry service responded with code %v", resp.StatusCode)
	}
	return nil
}

func DoShutdown(url string) error {
	request, err := http.NewRequest(http.MethodDelete, ServiceUrl, bytes.NewBuffer([]byte(url)))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "text/plain")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to shutdown service. shutdown service responded with code %v", resp.StatusCode)
	}
	return nil
}

func (p *providers) Update(pat patch) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, patchEntry := range pat.Added {
		if _, ok := p.services[patchEntry.Name]; !ok {
			p.services[patchEntry.Name] = make([]string, 0)
		}
		p.services[patchEntry.Name] = append(p.services[patchEntry.Name], patchEntry.URL)
		fmt.Printf("add service %s \n", patchEntry.Name)
	}

	for _, patchEntry := range pat.Removed {
		if providerUrls, ok := p.services[patchEntry.Name]; ok {
			for i := range providerUrls {
				if providerUrls[i] == patchEntry.URL {
					p.services[patchEntry.Name] = append(providerUrls[:i], providerUrls[i+1:]...)
				}
			}
			fmt.Printf("remove service %s \n", patchEntry.Name)
		}
	}
}

func (p providers) get(name ServiceName) (string, error) {
	providerUrls, ok := p.services[name]
	if !ok {
		return "", fmt.Errorf("No providers available for service %v", name)
	}
	idx := int(rand.Float32() * float32(len(providerUrls)))
	return providerUrls[idx], nil
}

func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)
}

type providers struct {
	services map[ServiceName][]string
	mutex    *sync.RWMutex
}

var prov = providers{
	services: make(map[ServiceName][]string),
	mutex:    new(sync.RWMutex),
}
