# Install

> Download the latest package

```shell
> go get github.com/redso/noq-roomq-go-sdk
```

# RoomQ Backend SDK - Go

The [RoomQ](https://www.noq.hk/en/roomq) Backend SDK is used for server-side integration to your server. It was developed with Go.

## High Level Logic

![The SDK Flow](https://raw.githubusercontent.com/redso/roomq.backend-sdk.nodejs/master/RoomQ-Backend-SDK-JS-high-level-logic-diagram.png)

1.  End user requests a page on your server
2.  The SDK verify if the request contain a valid ticket and in Serving state. If not, the SDK send him to the queue.
3.  End user obtain a ticket and wait in the queue until the ticket turns into Serving state.
4.  End user is redirected back to your website, now with a valid ticket
5.  The SDK verify if the request contain a valid ticket and in Serving state. End user stay in the requested page.
6.  The end user browses to a new page, and the SDK continue to check if the ticket is valid.

## How to integrate

### Prerequisite

To integrate with the SDK, you need to have the following information provided by RoomQ

1.  ROOM_ID
2.  ROOM_SECRET
3.  ROOMQ_TICKET_ISSUER
4.  ROOMQ_STATUS_API

### Major steps

To validate that the end user is allowed to access your site (has been through the queue) these steps are needed:

1.  Initialise RoomQ
2.  Determine if the current request page/path required to be protected by RoomQ
3.  Initialise Http Context Provider
4.  Validate the request
5.  If the end user should goes to the queue, set cache control
6.  Redirect user to queue

### Integration on specific path

It is recommended to integrate on the page/path which are selected to be provided. For the static files, e.g. images, css files, js files, ..., it is recommended to be skipped from the validation.
You can determine the requests type before pass it to the validation.

## Implementation Example

The following is an RoomQ integration example in Go.

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    NoQ_RoomQ "github.com/redso/noq-roomq-go-sdk"
    NoQ_RoomQ_Exception "github.com/redso/noq-roomq-go-sdk/Exception"
)

const (
    ROOM_ID             = "ROOM ID"
    ROOM_SECRET         = "ROOM SECRET"
    ROOMQ_TICKET_ISSUER = "TICKET ISSUER URL"
    ROOMQ_STATUS_API    = "STATUS API"
)

type Response map[string]interface{}

func Log(message ...interface{}) {
    log.Println(message...)
}

func getTicket(httpRes http.ResponseWriter, httpReq *http.Request) {
    defer errorHandler(httpRes)
    Log("Get Ticket")
    httpRes.Header().Set("Content-Type", "application/json")
    sessionID := httpReq.URL.Query().Get("sessionID")
    redirectURL := httpReq.URL.Query().Get("redirectURL")

    roomQ := NoQ_RoomQ.RoomQ(ROOM_ID, ROOM_SECRET, ROOMQ_TICKET_ISSUER, ROOMQ_STATUS_API, httpReq, true)
    validationResult := roomQ.Validate(httpReq, httpRes, redirectURL, sessionID)
    needRedirect := validationResult.NeedRedirect()
    if validationResult.NeedRedirect() {
        redirectURL = validationResult.GetRedirectURL()
    } else {
        // Proceed to main logic
        Log("Proceed to main logic")
    }
    val, _ := json.Marshal(Response{"redirect": needRedirect, "redirectURL": redirectURL})
    httpRes.Write(val)
}

func getServing(httpRes http.ResponseWriter, httpReq *http.Request) {
    defer errorHandler(httpRes)
    httpRes.Header().Set("Content-Type", "application/json")
    Log("Get serving")
    roomQ := NoQ_RoomQ.RoomQ(ROOM_ID, ROOM_SECRET, ROOMQ_TICKET_ISSUER, ROOMQ_STATUS_API, httpReq, true)
    if serving, err := roomQ.GetServing(); err != nil {
        httpRes.WriteHeader(http.StatusInternalServerError)
        val, _ := json.Marshal(Response{"error": err.Error()})
        httpRes.Write(val)
    } else {
        val, _ := json.Marshal(Response{"serving": serving})
        httpRes.Write(val)
    }
}

func extendTicket(httpRes http.ResponseWriter, httpReq *http.Request) {
    defer errorHandler(httpRes)
    httpRes.Header().Set("Content-Type", "application/json")
    Log("Extend")
    roomQ := NoQ_RoomQ.RoomQ(ROOM_ID, ROOM_SECRET, ROOMQ_TICKET_ISSUER, ROOMQ_STATUS_API, httpReq, true)
    if err := roomQ.Extend(httpRes, 60); err != nil {
        httpRes.WriteHeader(http.StatusInternalServerError)
        val, _ := json.Marshal(Response{"error": err.Error()})
        httpRes.Write(val)
    } else {
        if serving, err := roomQ.GetServing(); err != nil {
            httpRes.WriteHeader(http.StatusInternalServerError)
            val, _ := json.Marshal(Response{"error": err.Error()})
            httpRes.Write(val)
        } else {
            val, _ := json.Marshal(Response{"serving": serving})
            httpRes.Write(val)
        }
    }
}

func deleteTicket(httpRes http.ResponseWriter, httpReq *http.Request) {
    defer errorHandler(httpRes)
    httpRes.Header().Set("Content-Type", "application/json")
    Log("Extend")
    roomQ := NoQ_RoomQ.RoomQ(ROOM_ID, ROOM_SECRET, ROOMQ_TICKET_ISSUER, ROOMQ_STATUS_API, httpReq, true)
    if err := roomQ.DeleteServing(httpRes); err != nil {
        httpRes.WriteHeader(http.StatusInternalServerError)
        val, _ := json.Marshal(Response{"error": err.Error()})
        httpRes.Write(val)
    } else {
        val, _ := json.Marshal(Response{})
        httpRes.Write(val)
    }
}

func errorHandler(httpRes http.ResponseWriter) {
    if err := recover(); err != nil {
        switch e := err.(type) {
        case *NoQ_RoomQ_Exception.InvalidApiKeyException:
            http.Error(httpRes, e.Error(), http.StatusUnauthorized)
        case *NoQ_RoomQ_Exception.InvalidTokenException:
            http.Error(httpRes, e.Error(), http.StatusUnauthorized)
        case *NoQ_RoomQ_Exception.NotServingException:
            http.Error(httpRes, e.Error(), http.StatusNotFound)
        case *NoQ_RoomQ_Exception.QueueStoppedException:
            http.Error(httpRes, e.Error(), http.StatusGone)
        case *NoQ_RoomQ_Exception.ReachLimitException:
            http.Error(httpRes, e.Error(), http.StatusServiceUnavailable)
        default:
            Log(e)
            fmt.Fprint(httpRes, "{}")
        }
    }
}

func handleRequests() {
    http.HandleFunc("/api/roomq/get-ticket", getTicket)
    http.HandleFunc("/api/roomq/get-serving", getServing)
    http.HandleFunc("/api/roomq/extend-ticket", extendTicket)
    http.HandleFunc("/api/roomq/delete-ticket", deleteTicket)
    http.ListenAndServe("localhost:3000", nil)
}

func main() {
    Log("Web service started on port 3000")
    handleRequests()
}
```

### Ajax calls

RoomQ doesn't support validate ticket in Ajax calls yet.

### Browser / CDN cache

If your responses are cached on browser or CDN, the new requests will not process by RoomQ.
In general, for the page / path integrated with RoomQ, you are not likely to cache the responses on CDN or browser.

### Hash of URL

As hash of URL will not send to server, hash information will be lost.

## Version Guidance

| Version | Go dev pkg      | Go Version |
| ------- | :--------------- | ---------------------- |
| 1.x     | `github.com/redso/noq-roomq-go-sdk` | 1.17.1                    |
