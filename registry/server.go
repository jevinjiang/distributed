package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const ServerPort = ":3000"
const ServiceUrl = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []Registration
	mutex         *sync.RWMutex
}

var once sync.Once

func SetupRegistryService() {
	once.Do(func() {
		go reg.heartbeat(3 * time.Second)
	})
}

func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	err := r.sendRequiredServices(reg)
	r.notify(
		patch{
			Added: []patchEntry{
				{
					Name: reg.ServiceName,
					URL:  reg.ServiceURL,
				},
			},
		},
	)
	return err
}

func (r *registry) sendRequiredServices(reg Registration) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var p patch
	for _, serviceReg := range r.registrations {
		for _, reqService := range reg.RequiredServices {
			if serviceReg.ServiceName == reqService {
				p.Added = append(p.Added, patchEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}
		}
	}
	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}
	return nil
}

func (r *registry) remove(serviceUrl string) error {
	for i := range r.registrations {
		if r.registrations[i].ServiceURL == serviceUrl {
			r.notify(patch{
				Removed: []patchEntry{
					{
						Name: r.registrations[i].ServiceName,
						URL:  r.registrations[i].ServiceURL,
					},
				},
			})
			r.mutex.Lock()
			r.registrations = append(r.registrations[:i], r.registrations[i+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("Service at Url %s not found", serviceUrl)
}

func (r *registry) heartbeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup
		for _, reg := range r.registrations {
			wg.Add(1)
			go func(registration Registration) {
				defer wg.Done()
				success := true
				for attemps := 0; attemps < 3; attemps++ {
					resp, err := http.Get(registration.HeartbeatUrl)
					if err != nil {
						log.Println(err)
					} else if resp.StatusCode == http.StatusOK {
						log.Printf("Heartbeat check passed for %v", registration.ServiceName)
						if !success {
							r.add(registration)
						}
						break
					}
					log.Printf("Heartbeat check failed for %v", registration.ServiceName)
					if success {
						success = false
						r.remove(registration.ServiceURL)
					}
					time.Sleep(1 * time.Second)
				}
			}(reg)
			wg.Wait()
			time.Sleep(freq)
		}
	}
}

func (r *registry) sendPatch(p patch, url string) error {
	marshal, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(marshal))
	if err != nil {
		return err
	}
	return nil
}

func (r *registry) notify(fullPatch patch) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, reg := range r.registrations {
		go func(reg Registration) {
			for _, reqService := range reg.RequiredServices {
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}
				sendUpdate := false
				for _, added := range fullPatch.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}

				for _, removed := range fullPatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				if sendUpdate {
					err := r.sendPatch(p, reg.ServiceUpdateURL)
					if err != nil {
						fmt.Printf("Failed to notify service")
					}
				}
			}
		}(reg)
	}
}

var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.RWMutex),
}

type RegistryService struct{}

func (service RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received")
	switch r.Method {
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding Service: %v with URL: %s\n", r.ServiceName, r.ServiceURL)
		err = reg.add(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		url := string(payload)
		err = reg.remove(url)
		log.Printf("Remove Service with URL: %s\n", url)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
