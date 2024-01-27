package anova

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/xeipuuv/gojsonschema"
	"go-apo/anova/dto"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
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

type Client struct {
	conn               *websocket.Conn
	printMessageTraces bool

	stop chan bool

	inboundMessages chan dto.Message
}

func OptionPrintMessageTraces(client *Client) error {
	client.printMessageTraces = true
	return nil
}

func NewAnovaClient(firebaseRefreshToken string, options ...func(*Client) error) (client *Client, err error) {
	// Grab an ID token from Firebase Auth
	//TODO we probably need to save the refresh token in order to re-auth
	// if our WebSocket connection dies after more than an hour
	firebaseIdToken, err := getFirebaseIdToken(firebaseRefreshToken)
	if err != nil {
		return nil, err
	}

	conn, err := connectWebsocket(firebaseIdToken)

	client = &Client{
		conn:               conn,
		printMessageTraces: false,

		stop: make(chan bool),

		inboundMessages: make(chan dto.Message),
	}

	// Apply all options
	for _, option := range options {
		err := option(client)
		if err != nil {
			return nil, err
		}
	}

	go client.receiveMessages()

	return client, nil
}

func (client *Client) Close() {
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

func (client *Client) SetLamp(cookerID CookerID, on bool) error {
	command := dto.SetLampCommand{
		On: on,
	}
	_, err := client.SendCommand(cookerID, command)
	return err
}

// SetLampPreference sends a command to set the oven's default lamp state.
func (client *Client) SetLampPreference(cookerID CookerID, on bool) error {
	command := dto.SetLampPreferenceCommand{
		On: on,
	}
	_, err := client.SendCommand(cookerID, command)
	return err
}

// SetProbe sends a command to adjust the setpoint temperature of the temperature
// probe.
func (client *Client) SetProbe(cookerID CookerID, setpointCelsius float64) error {
	command := dto.SetProbeCommand{
		Setpoint: dto.TemperatureFromCelsius(setpointCelsius),
	}
	_, err := client.SendCommand(cookerID, command)
	return err
}

func (client *Client) StartCook(cookerID CookerID) error {
	command := dto.StartCookCommand{}
	_, err := client.SendCommand(cookerID, command)
	return err
}

// DisconnectOvenFromAccount sends a command to remove the oven from the current
// account. You will have to go through the Wi-Fi setup process all over again if
// no other accounts are paired with the oven.
func (client *Client) DisconnectOvenFromAccount(cookerID CookerID) error {
	command := dto.DisconnectCommand{}
	_, err := client.SendCommand(cookerID, command)
	return err
}

func (client *Client) SendCommand(cookerID CookerID, payload interface{}) (requestID dto.RequestID, err error) {
	payloadType := reflect.TypeOf(payload)
	messageType := dto.RequestTypeToMessageType[payloadType]
	if messageType == "" {
		return "", errors.New(fmt.Sprintf("missing message type for payload %s", payloadType))
	}

	requestID = dto.RequestID(uuid.New().String())

	command := dto.Command{
		ID:      dto.CookerID(cookerID),
		Type:    messageType,
		Payload: payload,
	}
	message := dto.Message{
		Command:   messageType,
		Payload:   command,
		RequestID: &requestID,
	}

	// Validate the entire command using the JSON Schema
	jsonLoader := gojsonschema.NewGoLoader(message)
	result, err := dto.OvenCommandSchema.Validate(jsonLoader)
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

	buf, err := json.Marshal(message)
	if err != nil {
		log.Printf("marshal error for message: %+v\n", message)
		return "", errors.New("failed to marshal message")
	}

	if client.printMessageTraces {
		log.Printf("send: %s\n", buf)
	}

	err = client.conn.WriteMessage(websocket.TextMessage, buf)
	if err != nil {
		log.Printf("send error \"%s\" for buffer \"%s\"\n", err, buf)
		return "", errors.New("failed to send message")
	}

	return requestID, nil
}

func (client *Client) receiveMessages() {
	for {
		select {
		case <-client.stop:
			break
		default:
		}

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
		var rawMessage dto.RawMessage
		err = dec.Decode(&rawMessage)
		if err != nil {
			log.Printf("error parsing JSON: %v\n", err)
			continue
		}

		// Decode the payload of the message
		if messageType := dto.MessageTypeToResponseType[rawMessage.Command]; messageType != nil {
			dec = json.NewDecoder(bytes.NewReader(rawMessage.Payload))
			dec.DisallowUnknownFields()

			decodedPayload := reflect.New(messageType).Interface()
			err = dec.Decode(decodedPayload)
			if err != nil {
				log.Printf("error parsing payload JSON for \"%s\": %v\npayload: %s\n", rawMessage.Command, err, message)
				continue
			}

			decodedMessage := dto.Message{
				RequestID: rawMessage.RequestID,
				Command:   rawMessage.Command,
				Payload:   decodedPayload,
			}

			// We only have a JSON Schema for EVENT_APO_STATE, so we can at
			// least perform a soft validation for those messages
			switch payload := decodedPayload.(type) {
			case *dto.ApoStateEvent:
				jsonLoader := gojsonschema.NewGoLoader(payload.State)
				result, err := dto.OvenStateSchema.Validate(jsonLoader)
				if err == nil && !result.Valid() {
					log.Printf("the oven state is not valid:\n")
					for _, desc := range result.Errors() {
						log.Printf("- %s\n", desc)
					}
				}
			}

			client.inboundMessages <- decodedMessage
		} else {
			log.Printf("error: unmapped message type \"%s\": %+v\n", rawMessage.Command, message)
		}
	}
}

type ErrClosedConnection struct{}

func (e ErrClosedConnection) Error() string {
	return "connection was closed"
}

func (client *Client) ReadMessage() (message dto.Message, err error) {
	select {
	case <-client.stop:
		return dto.Message{}, ErrClosedConnection{}

	case message = <-client.inboundMessages:
		return message, nil
	}
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

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("close failed: %s\n", err)
		}
	}(rsp.Body)
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
