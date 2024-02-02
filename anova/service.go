package anova

import (
	"errors"
	"fmt"
	"github.com/r3labs/diff/v3"
	"go-apo/anova/dto"
	"log/slog"
	"reflect"
	"strings"
)

type User struct {
	ConnectedToAlexa      bool
	ConnectedToGoogleHome bool
}

// Service wraps a Client and provides a streamlined state and interface for
// interacting with connected ovens.
type Service struct {
	User  *User
	Ovens map[CookerID]*Oven

	client *Client

	events chan Event
}

func NewService(client *Client) *Service {
	service := &Service{
		User:  nil,
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
				return
			}

			slog.Error("failed to read message from client", err)
			continue
		}

		switch payload := message.Payload.(type) {
		// Account properties
		case *dto.UserStateEvent:
			newUser := User{
				ConnectedToAlexa:      payload.IsConnectedToAlexa,
				ConnectedToGoogleHome: payload.IsConnectedToGoogleHome,
			}

			service.User = &newUser

		// New device was paired with the account
		case *dto.WifiAddedEvent:
			// We will process the device when we receive the updated
			// WifiListEvent
			slog.Debug("new wifi device added",
				"cookerID", payload.CookerID)

		// List of devices paired with the account
		case *dto.WifiListEvent:
			for _, oven := range *payload {
				cookerID := CookerID(oven.CookerID)
				if existingOven, exists := service.Ovens[cookerID]; exists {
					if oven.Name != existingOven.Name {
						oldName := existingOven.Name
						existingOven.Name = oven.Name
						service.events <- OvenRenamed{Oven: existingOven, OldName: oldName}
					}

					if oven.Type != existingOven.Type {
						slog.Warn("oven appeared with a different type",
							"previousOven", existingOven,
							"newOven", oven)
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
				slog.Error("received state for unknown oven",
					"cookerID", cookerID)
				continue
			}

			if oven.Type != payload.Type {
				slog.Warn("oven has different type in latest state",
					"previousOvenType", oven.Type,
					"newOvenType", payload.Type)
			}

			previousState := oven.State
			oven.State = &state
			oven.LastUpdate = state.UpdatedTimestamp

			if previousState != nil {
				diffFilter := diff.Filter(func(path []string, parent reflect.Type, field reflect.StructField) bool {
					return !((len(path) >= 2 && path[0] == "State" && path[1] == "ProcessedCommandIDs") ||
						(len(path) >= 1 && path[0] == "UpdatedTimestamp"))
				})
				changelog, err := diff.Diff(*previousState, state, diffFilter)
				if err != nil {
					slog.Error("diffing oven states failed", err)
				} else {
					slog.Debug(fmt.Sprintf("found %d differences in oven state", len(changelog)))
					for i, change := range changelog {
						fromValue := change.From
						if fromValue != nil && reflect.TypeOf(fromValue).Kind() == reflect.Pointer {
							fromValue = reflect.ValueOf(fromValue).Elem()
						}
						toValue := change.To
						if toValue != nil && reflect.TypeOf(toValue).Kind() == reflect.Pointer {
							toValue = reflect.ValueOf(toValue).Elem()
						}
						slog.Debug(fmt.Sprintf("change %d: %s %v %+v -> %+v",
							i+1, change.Type, strings.Join(change.Path, "."), fromValue, toValue))
					}
				}

				if len(changelog) > 0 {
					service.events <- OvenUpdated{Oven: oven, PreviousState: previousState}
				}
			} else {
				service.events <- OvenAdded{Oven: oven}
			}

		default:
			slog.Warn("skipping message",
				slog.Any("requestID", message.RequestID),
				slog.Any("payloadType", reflect.TypeOf(payload)),
				slog.Any("payload", payload))
		}
	}
}
