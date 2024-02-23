# 🥐 `apocli`

A command-line interface for interacting with the Anova Precision Oven (APO)
over the cloud.

## 📋 Features 

- [x] Connect to the Anova backend
- [ ] Design a nice wrapper around the backend that can be split into a library

- [ ] Create a useful CLI

- [x] Add URL handler for macOS

## ⚠️ Disclaimer

This tool is not, in any way, affiliated with Anova. The API endpoints may
stop working one day; your Anova account may be banned/restricted for
unauthorized access to internal systems; your oven may spontaneously combust.

Use at your own risk.

## 🔨 Building

`apocli` requires that you have Go 1.21 and `make` installed.

To build and install the CLI to your Go path:
```
make install-apocli
```

### URL Handler (macOS only)

To build and install the URL handler:
```
make install-urlhandler
```

## 🤖 Usage

With the Go path in your shell path, interact with `apocli` directly:
```
apocli -h
```

### Initial Setup

Before you can interact with the Anova backend, you must first configure your
Firebase refresh token. This initial setup is a little complicated, but only
needs to be done once:
1. Sign in to https://anovaculinary.io/ali/
2. Visit https://oven.anovaculinary.com/
3. Open the developer console:
    - On Chrome and Firefox, right-click the page and select "Inspect," then click "Console"
    - On Safari, first [enable developer options](https://developer.apple.com/documentation/safari-developer-tools/enabling-developer-features), then right-click the page and select "Inspect Element," then click "Console"
4. Paste the following JavaScript into the Console and execute it:
```
let openReq = window.indexedDB.open("firebaseLocalStorageDb"); openReq.onsuccess = function() { let db = openReq.result; let txn = db.transaction("firebaseLocalStorage", "readonly"); let storage = txn.objectStore("firebaseLocalStorage"); let getReq = storage.getAll(); getReq.onsuccess = function() { console.log('Run the following command in your terminal:\napocli config set firebase-refresh-token "' + getReq.result[0].value.stsTokenManager.refreshToken + '"'); }; };
```
5. In your terminal, run the command that it produced. It should look like this:
```
% apocli config set firebase-refresh-token "AMf...roNI"
FirebaseRefreshToken (string) = AMf...roNI
```

### URL Handler

The URL handler currently accepts URLs of the form:

```
apo://<cooker-id>/<command>?<params>
```

`<cooker-id>` can either be:
- the 16-hexadecimal character ID of your cooker, or
- `default` to fall back to your `apocli` config's `DefaultCookerId`
    - If you haven't set a default cooker ID, this will further fall back to the
      first oven to respond on your account.

Supported commands are:
- `start-cook`: starts a single-stage cook with these URL-encoded params:
  - `mode`: either `dry` (non-sous vide mode) or `wet` (sous vide mode)
  - `temp`: temperature setpoint in either Fahrenheit or Celsius, e.g. `212f` or `100c`
  - `steam`: steam percentage (relative humidity in wet mode), e.g. `10` for 10%
  - `elements`: comma-separated list of heating elements, e.g. `rear,bottom`
  - `timer` (optional): time to cook for, e.g. `1h20m` or `5m10s`
  - `trigger` (optional): trigger for starting the timer, one of:
    - `button`: require the oven's button to be manually pressed
    - `preheated`: start the timer once the oven is preheated
    - `none`: start the timer immediately, i.e. as the oven preheats
- `stop-cook`: stops the current cook

For example, to start pre-heating the default oven to 392 °F in non-sous vide mode
with  the rear element, 10% steam, and a 40-minute timer that starts when the oven's
button is pressed:
```
apo://default/start-cook?mode=dry&temp=392f&steam=10&elements=rear&timer=40m&trigger=button
```

And to stop the oven:
```
apo://default/stop-cook
```

## 🫡 Acknowledgements

This tool stands on the shoulders of many who previously reverse engineered
the Anova backend:

- [**github.com/bogd/anova-oven-api**](https://github.com/bogd/anova-oven-api): the most complete documentation of the API, plus a nice Go wrapper
- [**github.com/huangyq23/anova-oven-forwarder**](https://github.com/huangyq23/anova-oven-forwarder): a Grafana dashboard for visualizing data from the API
- [**github.com/andr83/hacs-anova-oven**](https://github.com/andr83/hacs-anova-oven): a controllable Home Assistant integration using the API

## 📜 License

`apocli` is released under the MIT License. See `LICENSE` for more details.

The Anova Precision Oven and all related content, including the JSON Schemas in
`schemas/`, are property of Anova Applied Electronics, Inc.

## 🔎 How It All Works

API access has been [a long-requested feature of the APO](https://community.anovaculinary.com/t/api-in-2021/28843).
As of February 2024, Anova has yet to announce any prioritization of a public
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