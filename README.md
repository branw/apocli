# `go-apo`

A small command-line tool for interacting with the Anova Precision Oven (APO) 
over the cloud.

## 📋 Features and Roadmap

- [x] Connect to the Anova backend
- [ ] Create a nice, synchronous client for interacting with the backend

- [ ] Set up a CI/CD workflow and generate binaries
- [ ] Expose the interface used by the CLI as a Go library

- [ ] Investigate options for non-cloud connectivity (e.g. if the device leaves
  any ports open after Wi-Fi setup is completed)
- [ ] Reverse engineer non-cooking related APIs (e.g. account setup and device
  pairing)

## ⚠️ Disclaimer

- This tool is not, in any way, affiliated with Anova. The API endpoints may
  stop working one day; your Anova account may be banned/restricted for
  unauthorized access to internal systems; your oven may spontaneously combust
  Use at your own risk.
- This is my first foray into Golang, so there's probably a lot of room for
  improvement.

## 🫡 Acknowledgements

This tool stands on the shoulders of many who previously reverse engineered
the app:

- [**github.com/bogd/anova-oven-api**](https://github.com/bogd/anova-oven-api)

## 🔨 Usage

To build and run (this isn't very useful yet):

```
go build
ANOVA_REFRESH_TOKEN="<your-refresh-token-here>" ANOVA_COOKER_ID="<your-cooker-id-here>" ./apo-cli
```

## 📜 License

`go-apo` is released under the MIT License. See `LICENSE` for more details.

The Anova Precision Oven and all related content, including the JSON Schemas in
`schemas/`, are all owned by Anova Applied Electronics, Inc.

## 🔎 How It All Works

API access has been [a long-requested feature of the APO](https://community.anovaculinary.com/t/api-in-2021/28843).
As of January 2024, Anova has yet to announce any prioritization of a public
API. Fortunately, they have not taken any efforts to obfuscate their internal
APIs either, and we can gleam everything we need to know from their Android
app and website:

- Account management and authentication is performed using [Firebase Auth](https://firebase.google.com/docs/auth)
  - Anova uses a Firebase API key of `AIzaSyB0VNqmJVAeR1fn_NbqqhwSytyMOZ_JO9c`
    across all platforms
  - After authenticating, Firebase produces a long-lived "refresh token" and
    an ephemeral, JWT-based "ID token." The refresh token is only invalidated
    when a "[major account change](https://firebase.google.com/docs/auth/admin/manage-sessions)"
    occurs, whereas the ID token expires after an hour.
    - [The refresh token can be repeatedly used to generate new ID tokens](https://firebase.google.com/docs/reference/rest/auth#section-refresh-token)
- All monitoring and interaction with the ovens is performed over an SSL
  WebSocket connection to `devices.anovaculinary.io`
  - The WebSocket is opened with these query parameters (e.g. `?token=...&`):
    - `token`: Firebase ID token
    - `supportedAccessories`: always `APO`
    - `platform`: always `android` (`ios` for the iOS app, and so on)
  - The connection is made with these HTTP headers:
    - `Sec-WebSocket-Protocol`: always `ANOVA_V2`
    - `User-Agent`: always `okhttp/4.9.2` (the Android app uses [OkHttp 4.9.2](https://github.com/square/okhttp/tree/parent-4.9.2))
- The WebSocket connection sends and receives messages in JSON format. These
  messages include request-response style "commands" with request IDs, as well
  as spontaneous notification "events" from the server.
  - Messages for both directions use an envelope with the following fields:
    - `command`: a string identifying the message, e.g. `CMD_APO_SET_LAMP`
    - `payload`: message body, either an object or array
    - `requestId`: a randomly-generated UUID in 8-4-4-4-12 format, e.g. `39c0e6f0-42e0-4d8b-9f88-3bc5aa9478f3`
      - Only present for command messages
  - Command messages are initiated by the client with a request, and then are
    acknowledged by the server with a `RESPONSE` message.
    - All command request messages have a `command` prefixed with `CMD_`
      (e.g. `CMD_APO_SET_LAMP`), or in (lower) camel case (e.g. `setLamp`).
      Either `command` format can be used.
    - Command request messages have an additional envelope on top of the
      payload:
      - `id`: 18-character hex string identifying the APO, a.k.a. the "cooker
        ID"
        - This value can be discovered through either the `EVENT_APO_WIFI_LIST`
          or `EVENT_APO_STATE` event messages. It can also be found under the
          settings screen of the Android app.
        - The cooker ID does not change after unpairing and re-pairing, or when
          moving the oven between accounts
        - Sending a command with an invalid/non-existent cooker ID does not
          seem to do anything. No `RESPONSE` is returned at all.
      - `type`: the same thing as the upper-level `command`
      - `payload` (yes, `message['payload']['payload']`): the actual message 
        body
  - Response objects 
  - Event messages are sent to the client asynchronous of any commands
    - All event messages have a `command` prefixed with `EVENT_`
    - They do not have an additional payload envelope
  - Upon connecting to the WebSocket, the following events are immediately
    sent to the client:
    - `EVENT_APO_WIFI_LIST`
    - `EVENT_USER_STATE`
    - `EVENT_APO_STATE` for every APO in the Wi-Fi list
      - This event will be continuously sent every 30 seconds for idle ovens
        and every 2 seconds for running ovens.
  - When a new oven is paired, an `EVENT_APO_WIFI_ADDED` event is sent,
    followed by `EVENT_APO_WIFI_LIST`, `EVENT_USER_STATE`, and `EVENT_APO_STATE`

### Inbound Messages (server to client)

| Type                             | Description |
|----------------------------------|-------------|
| `EVENT_APO_WIFI_ADDED`           |             |
| `EVENT_APO_WIFI_REMOVED`         |             |
| `EVENT_APO_WIFI_LIST`            |             |
| `EVENT_USER_STATE`               |             |
| `EVENT_APO_STATE`                |             |
| `EVENT_APO_WIFI_FIRMWARE_UPDATE` |             |
| `RESPONSE`                       |             |

### Outbound Messages (client to server)

#### "Oven Commands"

| Type                                                                | Description |
|---------------------------------------------------------------------|-------------|
| `AUTH_TOKEN_V2`                                                     |             |
| `CMD_APO_DISCONNECT`                                                |             |
| `CMD_APO_REGISTER_PUSH_TOKEN`                                       |             |
| `CMD_APO_NAME_WIFI_DEVICE`                                          |             |
| `CMD_APO_OTA`/`startFirmwareUpdate`                                 |             |
| `CMD_APO_GET_CONFIGURATION`/`getConfiguration`                      |             |
| `CMD_APO_SET_CONFIGURATION`/`setConfiguration`                      |             |
| `CMD_APO_HEALTHCHECK`                                               |             |
| `CMD_APO_REQUEST_DIAGNOSTIC`/`requestDiagnostic`                    |             |
| `CMD_APO_SET_BOILER_TIME`/`setBoilerTime`                           |             |
| `CMD_APO_SET_FAN`/`setFan`                                          |             |
| `CMD_APO_SET_HEATING_ELEMENTS`/`setHeatingElements`                 |             |
| `CMD_APO_SET_LAMP`/`setLamp`                                        |             |
| `CMD_APO_SET_LAMP_PREFERENCE`/`lampPreference`                      |             |
| `CMD_APO_SET_PROBE`/`setProbe`                                      |             |
| `CMD_APO_SET_REPORT_STATE_RATE`/`setReportStateRate`                |             |
| `CMD_APO_SET_REPORT_STATE_RATE_DEFAULT`/`setReportStateRateDefault` |             |
| `CMD_APO_SET_STEAM_GENERATORS`/`setSteamGenerators`                 |             |
| `CMD_APO_SET_TEMPERATURE_BULBS`/`setTemperatureBulbs`               |             |
| `CMD_APO_SET_TEMPERATURE_UNIT`/`setTemperatureUnit`                 |             |
| `CMD_APO_SET_TIMER`/`setTimer`                                      |             |
| `CMD_APO_SET_VENT`/`setVent`                                        |             |
| `CMD_APO_SET_METADATA`                                              |             |
| `CMD_APO_SET_TIME_ZONE`                                             |             |
| `CMD_APO_START`/`startCook`                                         |             |
| `CMD_APO_STOP`/`stopCook`                                           |             |
| `CMD_APO_START_STAGE`/`startStage`                                  |             |
| `CMD_APO_UPDATE_COOK_STAGE`/`updateCookStage`                       |             |
| `CMD_APO_UPDATE_COOK_STAGES`/`updateCookStages`                     |             |
| `CMD_APO_START_DESCALE`/`startDescale`                              |             |
| `CMD_APO_ABORT_DESCALE`/`abortDescale`                              |             |
| `CMD_APO_START_LIVE_STREAM`                                         |             |
| `CMD_APO_STOP_LIVE_STREAM`                                          |             |

### "Multi-User Commands"

| Type                        | Description |
|-----------------------------|-------------|
| `CMD_USER_PAIR_ALEXA`       |             |
| `CMD_USER_UNPAIR_ALEXA`     |             |
| `CMD_LIST_USERS`            |             |
| `CMD_GENERATE_NEW_PAIRING`  |             |
| `CMD_ADD_USER_WITH_PAIRING` |             |
