package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/r3labs/diff/v3"
	"github.com/xeipuuv/gojsonschema"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// From Android APK "com.anovaculinary.anovaoven" 1.1.7 (2023-12-09)
const (
	FirebaseAPIKey                   = "AIzaSyB0VNqmJVAeR1fn_NbqqhwSytyMOZ_JO9c"
	AnovaDevicesEndpoint             = "wss://devices.anovaculinary.io/"
	AnovaDevicesSupportedAccessories = "APO"
	AnovaDevicesPlatform             = "android"
	AnovaDevicesWebSocketProtocol    = "ANOVA_V2"
	AnovaDevicesUserAgent            = "okhttp/4.9.2"
)

// JSON Schema for "EVENT_APO_STATE" messages
//
//go:embed schemas/oven_state_schema.json
var ovenStateSchemaJson string

// JSON Schema for outbound command messages
//
//go:embed schemas/oven_command_schema.json
var ovenCommandSchemaJson string

// JSON Schema for outbound multi-user command messages
//
//go:embed schemas/multi_user_command_schema.json
var multiUserCommandSchemaJson string

type Oven struct {
	CookerID CookerID
	Name     string
	Type     string
	State    *OvenStateV1
}

type AnovaClient struct {
	conn               *websocket.Conn
	printMessageTraces bool

	stop chan bool

	pendingRequests  map[RequestID]Message
	outboundMessages chan Message
	inboundMessages  chan Message

	ovenStateSchema   *gojsonschema.Schema
	ovenCommandSchema *gojsonschema.Schema

	Ovens     map[CookerID]*Oven
	UserState *EventUserState
}

func (client *AnovaClient) Close() {
	client.stop <- true

	err := client.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Printf("error on write close: %s\n", err)
		return
	}

	select {
	case <-time.After(time.Second):
	}
}

func OptionPrintMessageTraces(client *AnovaClient) error {
	client.printMessageTraces = true
	return nil
}

func NewAnovaClient(firebaseRefreshToken string, options ...func(*AnovaClient) error) (client *AnovaClient, err error) {
	// Grab an ID token from Firebase Auth
	//TODO we probably need to save the refresh token in order to re-auth
	// if our WebSocket connection dies after more than an hour
	firebaseIdToken, err := getFirebaseIdToken(firebaseRefreshToken)
	if err != nil {
		return nil, err
	}

	conn, err := connectWebsocket(firebaseIdToken)

	// Load the JSON Schemas so we can validate our messages
	ovenStateSchema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(ovenStateSchemaJson))
	if err != nil {
		return nil, err
	}
	ovenCommandSchema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(ovenCommandSchemaJson))
	if err != nil {
		return nil, err
	}

	client = &AnovaClient{
		conn:               conn,
		printMessageTraces: false,

		stop: make(chan bool),

		pendingRequests:  make(map[RequestID]Message),
		outboundMessages: make(chan Message),
		inboundMessages:  make(chan Message),

		ovenStateSchema:   ovenStateSchema,
		ovenCommandSchema: ovenCommandSchema,

		Ovens:     make(map[CookerID]*Oven),
		UserState: nil,
	}

	// Apply all options
	for _, option := range options {
		err := option(client)
		if err != nil {
			return nil, err
		}
	}

	go client.run()
	go client.receiveMessages()

	return client, nil
}

// Retrieves a one-hour session token from Firebase. The returned JWT contains
// account info, including an email and OAuth sign-in identity info.
// See https://firebase.google.com/docs/auth/admin/manage-sessions
func getFirebaseIdToken(refreshToken string) (string, error) {
	authURL, err := url.Parse("https://securetoken.googleapis.com/v1/token")
	if err != nil {
		return "", err
	}
	authURL.RawQuery = url.Values{
		"key": {FirebaseAPIKey},
	}.Encode()

	bodyValues := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	rsp, err := http.PostForm(authURL.String(), bodyValues)
	if err != nil {
		return "", err
	}

	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	var jsonResult = make(map[string]string)
	if err = json.Unmarshal(body, &jsonResult); err != nil {
		return "", err
	}

	accessToken, ok := jsonResult["access_token"]
	if !ok {
		return "", errors.New("missing access_token in response")
	}

	return accessToken, nil
}

func connectWebsocket(firebaseIdToken string) (conn *websocket.Conn, err error) {
	devicesURL, err := url.Parse(AnovaDevicesEndpoint)
	if err != nil {
		return nil, err
	}
	devicesURL.RawQuery = url.Values{
		"token":                {firebaseIdToken},
		"supportedAccessories": {AnovaDevicesSupportedAccessories},
		"platform":             {AnovaDevicesPlatform},
	}.Encode()

	headers := http.Header{
		"Sec-WebSocket-Protocol": {AnovaDevicesWebSocketProtocol},
		"User-Agent":             {AnovaDevicesUserAgent},
	}
	conn, _, err = websocket.DefaultDialer.Dial(devicesURL.String(), headers)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (client *AnovaClient) run() {
	for {
		select {
		case <-client.stop:
			break

		case message := <-client.outboundMessages:
			buf, err := json.Marshal(message)
			if err != nil {
				log.Printf("marshal error for message: %+v\n", message)
				continue
			}

			if message.RequestID == nil {
				log.Printf("outbound message has nil request ID: %+v\n", message)
			} else {
				if _, present := client.pendingRequests[*message.RequestID]; present {
					log.Printf("outbound message with this request ID has already been sent: %+v\n", message)
					continue
				}

				client.pendingRequests[*message.RequestID] = message
			}

			if client.printMessageTraces {
				log.Printf("send: %s\n", buf)
			}

			err = client.conn.WriteMessage(websocket.TextMessage, buf)
			if err != nil {
				log.Printf("send error \"%s\" for buffer \"%s\"\n", err, buf)
				continue
			}

		case message := <-client.inboundMessages:
			switch payload := message.Payload.(type) {
			case *EventApoWifiList:
				for _, oven := range *payload {
					if existingOven, exists := client.Ovens[oven.CookerID]; exists {
						if oven.Name != existingOven.Name || oven.Type != existingOven.Type {
							log.Printf("oven appeared with a different name/type. previous: %+v\nnew: %+v\n",
								existingOven, oven)
						}
					} else {
						client.Ovens[oven.CookerID] = &Oven{
							CookerID: oven.CookerID,
							Name:     oven.Name,
							Type:     oven.Type,
						}
					}
				}

			case *EventUserState:
				client.UserState = payload

			case *EventApoState:
				state := payload.State

				// Validate against the JSON Schema just for fun
				jsonLoader := gojsonschema.NewGoLoader(state)
				result, err := client.ovenStateSchema.Validate(jsonLoader)
				if err == nil && !result.Valid() {
					log.Printf("the oven state is not valid:\n")
					for _, desc := range result.Errors() {
						log.Printf("- %s\n", desc)
					}
				}

				oven, exists := client.Ovens[payload.CookerID]
				if !exists {
					log.Printf("received state for non-existent oven with ID \"%s\"\n", payload.CookerID)
					continue
				}

				if oven.Type != payload.Type {
					log.Printf("oven has different type in latest state. previous: %s. new: %s.\n",
						oven.Type, payload.Type)
				}

				if oven.State != nil {
					diffFilter := diff.Filter(func(path []string, parent reflect.Type, field reflect.StructField) bool {
						return !((len(path) >= 2 && path[0] == "State" && path[1] == "ProcessedCommandIDs") ||
							(len(path) >= 1 && path[0] == "UpdatedTimestamp"))
					})
					changelog, err := diff.Diff(*oven.State, state, diffFilter)
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
				}

				oven.State = &state

			case *Response:
				requestMessage, present := client.pendingRequests[*message.RequestID]
				if !present {
					log.Printf("received response for unknown request ID: %+v\n", message)
				}
				delete(client.pendingRequests, *message.RequestID)

				if payload.Status != ResponseStatusOk {
					log.Printf("received negative response \"%+v\" for request \"%+v\"\n", message, requestMessage)
				} else {
					log.Printf("ack'd %s\n", *message.RequestID)
				}

			default:
				log.Printf("unhandled inbound message with type \"%+v\"\n", reflect.TypeOf(message.Payload))
			}
		}
	}
}

func (client *AnovaClient) QueueCommand(cookerID CookerID, payload interface{}) (requestID RequestID, err error) {
	payloadType := reflect.TypeOf(payload)
	messageType := requestTypeToMessageType[payloadType]
	if messageType == "" {
		return "", errors.New(fmt.Sprintf("missing message type for payload %s", payloadType))
	}

	requestID = RequestID(uuid.New().String())

	command := Command{
		ID:      cookerID,
		Type:    messageType,
		Payload: payload,
	}
	message := Message{
		Command:   messageType,
		Payload:   command,
		RequestID: &requestID,
	}

	// Validate the entire command using the JSON Schema
	jsonLoader := gojsonschema.NewGoLoader(message)
	result, err := client.ovenCommandSchema.Validate(jsonLoader)
	if err != nil {
		return "", err
	}

	if !result.Valid() {
		fmt.Printf("the command payload is not valid:\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return "", errors.New("invalid payload")
	}

	client.outboundMessages <- message

	return requestID, nil
}

func (client *AnovaClient) receiveMessages() {
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if errors.Is(err, websocket.ErrCloseSent) {
				break
			}

			log.Printf("read failed: %s\n", err)
			return
		}

		if client.printMessageTraces {
			log.Printf("recv: %s\n", message)
		}

		// Decode the first layer of the message
		dec := json.NewDecoder(bytes.NewReader(message))
		dec.DisallowUnknownFields()
		var rawMessage RawMessage
		err = dec.Decode(&rawMessage)
		if err != nil {
			log.Printf("error parsing JSON: %v\n", err)
			continue
		}

		// Decode the payload of the message
		if messageType := messageTypeToResponseType[rawMessage.Command]; messageType != nil {
			dec = json.NewDecoder(bytes.NewReader(rawMessage.Payload))
			dec.DisallowUnknownFields()

			decodedPayload := reflect.New(messageType).Interface()
			err = dec.Decode(decodedPayload)
			if err != nil {
				log.Printf("error parsing payload JSON for \"%s\": %v\npayload: %s\n", rawMessage.Command, err, message)
				continue
			}

			decodedMessage := Message{
				RequestID: rawMessage.RequestID,
				Command:   rawMessage.Command,
				Payload:   decodedPayload,
			}

			client.inboundMessages <- decodedMessage
		} else {
			log.Printf("error: unmapped message type \"%s\": %+v\n", rawMessage.Command, message)
		}
	}
}
