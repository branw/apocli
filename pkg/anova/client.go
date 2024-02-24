package anova

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/xeipuuv/gojsonschema"
	"go-apo/pkg/anova/dto"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"
)

// Anova-specific constants
// From Android APK "com.anovaculinary.anovaoven" 1.1.7 (2023-12-09)
const (
	firebaseAPIKey                   = "AIzaSyB0VNqmJVAeR1fn_NbqqhwSytyMOZ_JO9c"
	anovaDevicesEndpoint             = "wss://devices.anovaculinary.io/"
	anovaDevicesSupportedAccessories = "APO"
	anovaDevicesPlatform             = "android"
	anovaDevicesWebSocketProtocol    = "ANOVA_V2"
	anovaDevicesUserAgent            = "okhttp/4.9.2"
	// cookIdPrefix is the prefix added to all cook IDs and stage IDs
	cookIdPrefix = "android-"
)

const (
	// requestAcknowledgementTimeout is the timeout to wait for a RESPONSE message
	// after sending a request
	requestAcknowledgementTimeout = 5 * time.Second
)

type Client struct {
	conn      *websocket.Conn
	connMutex sync.Mutex

	printMessageTraces bool

	stop    chan bool
	stopped bool

	inboundMessages  chan dto.Message
	requestResponses map[dto.RequestID]chan map[string]interface{}
}

func OptionPrintMessageTraces(client *Client) error {
	client.printMessageTraces = true
	return nil
}

func NewClient(firebaseRefreshToken string, options ...func(*Client) error) (client *Client, err error) {
	// Grab an ID token from Firebase Auth
	//TODO we probably need to save the refresh token in order to re-auth
	// if our WebSocket connection dies after more than an hour
	firebaseIdToken, err := getFirebaseIdToken(firebaseRefreshToken)
	if err != nil {
		return nil, err
	}

	conn, err := connectWebsocket(firebaseIdToken)

	client = &Client{
		conn: conn,

		printMessageTraces: false,

		stop:    make(chan bool),
		stopped: false,

		inboundMessages:  make(chan dto.Message, 100),
		requestResponses: make(map[dto.RequestID]chan map[string]interface{}),
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
	if client.stopped {
		return
	}

	client.stop <- true
	client.stopped = true

	err := client.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		slog.Error("failed to write close message", slog.Any("err", err))
		return
	}

	select {
	case <-time.After(time.Second):
	}
}

// ErrRequestNotAcknowledged indicates that we did not receive a response for a
// command request. This can happen if there are networking issues, and also if
// the command was sent to a nonexistent cooker ID.
type ErrRequestNotAcknowledged struct {
	RequestID dto.RequestID
}

func (e ErrRequestNotAcknowledged) Error() string {
	return fmt.Sprintf("request \"%s\" was not acknowledged", e.RequestID)
}

// ErrRequestFailed indicates that a request command returned an error or other
// unsuccessful response.
type ErrRequestFailed struct {
	RequestID    dto.RequestID
	ErrorMessage string
}

func (e ErrRequestFailed) Error() string {
	return fmt.Sprintf("request \"%s\" failed with error: %s\n", e.RequestID, e.ErrorMessage)
}

func (client *Client) SendCommand(cookerID CookerID, payload interface{}) (requestID dto.RequestID, data map[string]interface{}, err error) {
	payloadType := reflect.TypeOf(payload)
	messageType := dto.RequestTypeToMessageType[payloadType]
	if messageType == "" {
		return "", nil, errors.New(fmt.Sprintf("missing message type for payload %s", payloadType))
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
	commandResult, err := dto.OvenCommandSchema.Validate(jsonLoader)
	if err != nil {
		return "", nil, err
	}
	multiUserResult, err := dto.MultiUserCommandSchema.Validate(jsonLoader)
	if err != nil {
		return "", nil, err
	}
	if !commandResult.Valid() && !multiUserResult.Valid() {
		slog.Error("outbound message payload is not valid",
			slog.Any("message", message),
			slog.Any("commandValidationErrors", commandResult.Errors()),
			slog.Any("multiUserValidationErrors", multiUserResult.Errors()))
		return "", nil, errors.New("invalid payload")
	}

	buf, err := json.Marshal(message)
	if err != nil {
		slog.Error("JSON marshal error for message",
			slog.Any("err", err),
			"message", message)
		return "", nil, errors.New("failed to marshal message")
	}

	if client.printMessageTraces {
		slog.Debug("send",
			slog.String("buf", string(buf)))
	}

	client.requestResponses[requestID] = make(chan map[string]interface{}, 1)
	defer delete(client.requestResponses, requestID)

	client.connMutex.Lock()
	{
		defer client.connMutex.Unlock()

		err = client.conn.WriteMessage(websocket.TextMessage, buf)
		if err != nil {
			slog.Error("failed to write message",
				slog.Any("err", err),
				"buffer", buf)
			return "", nil, errors.New("failed to send message")
		}
	}

	// Block until we receive an acknowledgement
	select {
	case response := <-client.requestResponses[requestID]:
		if response["status"].(string) == string(dto.ResponseStatusOk) {
			return requestID, response, nil
		}

		var errorMessage string
		if response["error"] != nil {
			errorMessage = response["error"].(string)
		} else {
			errorMessage = "(unknown error)"
		}
		return "", nil, ErrRequestFailed{RequestID: requestID, ErrorMessage: errorMessage}

	case <-time.After(requestAcknowledgementTimeout):
		return "", nil, ErrRequestNotAcknowledged{RequestID: requestID}
	}
}

func (client *Client) receiveMessages() {
	for {
		if client.stopped {
			return
		}

		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if errors.Is(err, websocket.ErrCloseSent) {
				return
			}

			slog.Error("failed to read from websocket",
				slog.Any("err", err))
			client.Close()
			return
		}

		if client.printMessageTraces {
			slog.Debug("recv",
				slog.String("message", string(message)))
		}

		// Decode the first layer of the message
		dec := json.NewDecoder(bytes.NewReader(message))
		dec.DisallowUnknownFields()
		var rawMessage dto.RawMessage
		err = dec.Decode(&rawMessage)
		if err != nil {
			slog.Error("error parsing JSON",
				slog.Any("err", err))
			continue
		}

		// Decode the payload of the message
		if rawMessage.Command == "RESPONSE" {
			if rawMessage.RequestID == nil {
				slog.Warn("received response with no request ID; ignoring")
				break
			}

			var payload map[string]interface{}
			err = json.Unmarshal(rawMessage.Payload, &payload)
			if err != nil {
				slog.Error("error unmarshalling response",
					slog.Any("err", err))
				continue
			}

			decodedMessage := dto.Message{
				RequestID: rawMessage.RequestID,
				Command:   rawMessage.Command,
				Payload:   payload,
			}

			status, exists := payload["status"]
			if !exists {
				slog.Error("response did not contain status field")
				continue
			}

			if status == dto.ResponseStatusOk {

			}

			// Unblock the synchronous message sender
			requestID := *decodedMessage.RequestID
			ch, exists := client.requestResponses[requestID]
			if !exists {
				slog.Warn("received response for unknown request ID",
					"requestID", requestID)
				break
			}
			ch <- payload

			// Still send the response to any listeners.
			client.inboundMessages <- decodedMessage
		} else if messageType := dto.MessageTypeToResponseType[rawMessage.Command]; messageType != nil {
			dec = json.NewDecoder(bytes.NewReader(rawMessage.Payload))
			dec.DisallowUnknownFields()

			decodedPayload := reflect.New(messageType).Interface()
			err = dec.Decode(decodedPayload)
			if err != nil {
				slog.Error("error parsing payload JSON",
					slog.Any("err", err),
					slog.String("command", string(rawMessage.Command)),
					slog.String("message", string(message)))
				continue
			}

			decodedMessage := dto.Message{
				RequestID: rawMessage.RequestID,
				Command:   rawMessage.Command,
				Payload:   decodedPayload,
			}

			switch payload := decodedPayload.(type) {
			case *dto.ApoStateEvent:
				// We only have a JSON Schema for EVENT_APO_STATE, so we can at
				// least perform a soft validation for those messages
				jsonLoader := gojsonschema.NewGoLoader(payload.State)
				result, err := dto.OvenStateSchema.Validate(jsonLoader)
				if err == nil && !result.Valid() {
					slog.Warn("the oven state is not valid",
						slog.Any("errors", result.Errors()))
				}
			}

			client.inboundMessages <- decodedMessage
		} else {
			slog.Error("unmapped message type",
				"command", rawMessage.Command,
				"message", message)
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
	if len(refreshToken) == 0 {
		return "", fmt.Errorf("invalid Firebase refresh token \"%s\"", refreshToken)
	}

	authURL, err := url.Parse("https://securetoken.googleapis.com/v1/token")
	if err != nil {
		return "", err
	}
	authURL.RawQuery = url.Values{
		"key": {firebaseAPIKey},
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
		err = Body.Close()
		if err != nil {
			slog.Error("failed to close token connection",
				slog.Any("err", err))
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
	devicesURL, err := url.Parse(anovaDevicesEndpoint)
	if err != nil {
		return nil, err
	}
	devicesURL.RawQuery = url.Values{
		"token":                {firebaseIdToken},
		"supportedAccessories": {anovaDevicesSupportedAccessories},
		"platform":             {anovaDevicesPlatform},
	}.Encode()

	headers := http.Header{
		"Sec-WebSocket-Protocol": {anovaDevicesWebSocketProtocol},
		"User-Agent":             {anovaDevicesUserAgent},
	}
	conn, _, err = websocket.DefaultDialer.Dial(devicesURL.String(), headers)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (client *Client) SetLamp(cookerID CookerID, on bool) error {
	command := dto.SetLampCommand{
		On: on,
	}
	_, _, err := client.SendCommand(cookerID, command)
	return err
}

// SetLampPreference sends a command to set the oven's default lamp state.
func (client *Client) SetLampPreference(cookerID CookerID, on bool) error {
	command := dto.SetLampPreferenceCommand{
		On: on,
	}
	_, _, err := client.SendCommand(cookerID, command)
	return err
}

// SetProbe sends a command to adjust the setpoint temperature of the temperature
// probe.
func (client *Client) SetProbe(cookerID CookerID, setpointCelsius float64) error {
	command := dto.SetProbeCommand{
		Setpoint: dto.NewTemperatureFromCelsius(setpointCelsius),
	}
	_, _, err := client.SendCommand(cookerID, command)
	return err
}

//func (client *Client) StartCook(cookerID CookerID) error {
//	command := dto.StartCookCommand{}
//	_, _, err := client.SendCommand(cookerID, command)
//	return err
//}

// DisconnectOvenFromAccount sends a command to remove the oven from the current
// account. You will have to go through the Wi-Fi setup process all over again if
// no other accounts are paired with the oven.
func (client *Client) DisconnectOvenFromAccount(cookerID CookerID) error {
	command := dto.DisconnectCommand{}
	_, _, err := client.SendCommand(cookerID, command)
	return err
}

func (client *Client) SetName(cookerID CookerID, name string) error {
	command := dto.NameWifiDeviceCommand{
		Name: name,
	}
	_, _, err := client.SendCommand(cookerID, command)
	return err
}

// GeneratePairingCode sends a command to generate a new pairing code for an
// oven. The pairing code is a 24-hour-long JWT that can be used from another
// account via AddUserWithPairingCode.
func (client *Client) GeneratePairingCode(cookerID CookerID) (pairingCode string, err error) {
	command := dto.GenerateNewPairingCode{}
	_, data, err := client.SendCommand(cookerID, command)
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", errors.New("command returned no response")
	}
	pairingCode, exists := data["data"].(string)
	if !exists {
		return "", errors.New("command returned no data")
	}
	return pairingCode, nil
}

// AddUserWithPairingCode sends a command to add an oven to the current account.
// A pairing code can be generated by calling GeneratePairingCode from a separate
// account.
func (client *Client) AddUserWithPairingCode(pairingCode string) error {
	command := dto.AddUserWithPairingCode{
		Data: pairingCode,
	}
	// Intentionally no cooker ID -- the oven isn't accessible from this account yet
	_, _, err := client.SendCommand("", command)
	return err
}

// ListUsersForDevice sends a command that returns a list of user IDs that have
// access to an oven
func (client *Client) ListUsersForDevice(cookerID CookerID) (userIds []string, err error) {
	command := dto.ListUsersForDevice{}
	_, data, err := client.SendCommand(cookerID, command)
	if err != nil {
		return nil, err
	}

	for _, userId := range data["userIds"].([]interface{}) {
		userIds = append(userIds, userId.(string))
	}
	return userIds, nil
}

func generateRandomCookUuid() string {
	return cookIdPrefix + uuid.New().String()
}

func (client *Client) StartCook(cookerID CookerID, cookId string, stages []dto.CookingStage) error {
	command := dto.StartCookCommandV1{
		CookID: cookId,
		Stages: stages,
	}
	_, _, err := client.SendCommand(cookerID, command)
	return err
}

// This really should be taking dto.OvenStage, but...
func (client *Client) UpdateCookStage(cookerID CookerID, stage dto.CookingStage) error {
	command := dto.UpdateCookStageCommand(stage)
	_, _, err := client.SendCommand(cookerID, command)
	return err
}

func (client *Client) UpdateCookStages(cookerID CookerID, stages []dto.CookingStage) error {
	command := dto.UpdateCookStagesCommand{
		Stages: stages,
	}
	_, _, err := client.SendCommand(cookerID, command)
	return err
}

func (client *Client) StopCook(cookerID CookerID) error {
	command := dto.StopCookCommand{}
	_, _, err := client.SendCommand(cookerID, command)
	return err
}
