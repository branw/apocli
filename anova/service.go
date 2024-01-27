package anova

import (
	"errors"
	"github.com/r3labs/diff/v3"
	"go-apo/anova/dto"
	"log"
	"reflect"
	"strings"
)

// Service wraps a Client and provides a streamlined state and interface for
// interacting with connected ovens.
type Service struct {
	Ovens map[CookerID]*Oven

	client *Client

	events chan Event
}

func NewService(client *Client) *Service {
	service := &Service{
		Ovens: make(map[CookerID]*Oven),

		client: client,

		events: make(chan Event),
	}

	go service.receiveMessages()

	return service
}

func (service *Service) ReadEvent() Event {
	select {
	case event := <-service.events:
		return event
	}
}

func (service *Service) receiveMessages() {
	for {
		message, err := service.client.ReadMessage()
		if err != nil {
			if errors.Is(err, ErrClosedConnection{}) {
				service.events <- ServiceStopped{}
				break
			}

			log.Printf("read error: %s\n", err)
			continue
		}

		switch payload := message.Payload.(type) {
		// New device was paired with the account
		case *dto.WifiAddedEvent:
			// We will process the device when we receive the updated
			// WifiListEvent
			log.Printf("new wifi device added with ID \"%s\"\n", payload.CookerID)

		// List of devices paired with the account
		case *dto.WifiListEvent:
			for _, oven := range *payload {
				cookerID := CookerID(oven.CookerID)
				if existingOven, exists := service.Ovens[cookerID]; exists {
					if oven.Name != existingOven.Name || oven.Type != existingOven.Type {
						log.Printf("oven appeared with a different name/type\nprevious: %+v\nnew: %+v\n",
							existingOven, oven)
					}
				} else {
					// Create an oven object, but defer sending an event until
					// we have the full state of the oven
					service.Ovens[cookerID] = &Oven{
						CookerID: CookerID(oven.CookerID),
						Name:     oven.Name,
						Type:     oven.Type,

						client: service.client,
					}
				}
			}

		// State of a paired deice
		case *dto.ApoStateEvent:
			cookerID := CookerID(payload.CookerID)
			state := payload.State

			oven, exists := service.Ovens[cookerID]
			if !exists {
				log.Printf("received state for non-existent oven with ID \"%s\"\n", cookerID)
				continue
			}

			if oven.Type != payload.Type {
				log.Printf("oven has different type in latest state. previous: %s. new: %s.\n",
					oven.Type, payload.Type)
			}

			previousState := oven.State
			oven.State = &state

			if previousState != nil {
				diffFilter := diff.Filter(func(path []string, parent reflect.Type, field reflect.StructField) bool {
					return !((len(path) >= 2 && path[0] == "State" && path[1] == "ProcessedCommandIDs") ||
						(len(path) >= 1 && path[0] == "UpdatedTimestamp"))
				})
				changelog, err := diff.Diff(*previousState, state, diffFilter)
				if err != nil {
					log.Printf("diffing oven states failed: %s\n", err)
				} else {
					log.Printf("found %d differences in oven state\n", len(changelog))
					for i, change := range changelog {
						fromValue := change.From
						if fromValue != nil && reflect.TypeOf(fromValue).Kind() == reflect.Pointer {
							fromValue = reflect.ValueOf(fromValue).Elem()
						}
						toValue := change.To
						if toValue != nil && reflect.TypeOf(toValue).Kind() == reflect.Pointer {
							toValue = reflect.ValueOf(toValue).Elem()
						}
						log.Printf("change %d: %s %v %+v -> %+v\n",
							i+1, change.Type, strings.Join(change.Path, "."), fromValue, toValue)
					}
				}

				if len(changelog) > 0 {
					service.events <- OvenUpdated{Oven: oven, PreviousState: previousState}
				}
			} else {
				service.events <- OvenAdded{Oven: oven}
			}

		default:
			log.Printf("skipping message %s: %+v\n", reflect.TypeOf(payload), payload)
		}
	}
}
