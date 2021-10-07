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
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    NoQ_RoomQ "github.com/koktszhozelca/noq-roomq-go-sdk"
    NoQ_RoomQ_Exception "github.com/koktszhozelca/noq-roomq-go-sdk/Exception"
)

const (
    ROOM_ID             = "ROOM ID"
    ROOM_SECRET         = "ROOM SECRET"
    ROOMQ_TICKET_ISSUER = "TICKET ISSUER URL"
    ROOMQ_STATUS_API    = "STATUS API"
)

type Response gin.H

func Log(message ...interface{}) {
    log.Println(message)
}

func getTicket(ctx *gin.Context) {
    defer errorHandler(ctx)
    Log("Get Ticket")
    var resp Response
    sessionID := ctx.Query("sessionID")
    redirectURL := ctx.Query("redirectURL")
    roomQ := NoQ_RoomQ.RoomQ(ROOM_ID, ROOM_SECRET, ROOMQ_TICKET_ISSUER, ROOMQ_STATUS_API, ctx, true)
    // Check if the request has valid ticket
    // If "session id" is null, SDK will generate UUID as "session id"
    validationResult := roomQ.Validate(ctx, redirectURL, sessionID)
    needRedirect := validationResult.NeedRedirect()
    if validationResult.NeedRedirect() {
        redirectURL = validationResult.GetRedirectURL()
    } else {
        // Proceed to main logic
    }
    Log(resp)
    ctx.IndentedJSON(http.StatusOK, Response{
        "redirect":    needRedirect,
        "redirectURL": redirectURL,
    })
}

func getServing(ctx *gin.Context) {
    defer errorHandler(ctx)
    Log("Get serving")
    roomQ := NoQ_RoomQ.RoomQ(ROOM_ID, ROOM_SECRET, ROOMQ_TICKET_ISSUER, ROOMQ_STATUS_API, ctx, true)
    if serving, err := roomQ.GetServing(ctx); err != nil {
        ctx.IndentedJSON(http.StatusInternalServerError, Response{"error": err.Error()})
    } else {
        ctx.IndentedJSON(http.StatusOK, Response{"serving": serving})
    }
}

func extendTicket(ctx *gin.Context) {
    defer errorHandler(ctx)
    Log("Extend")
    roomQ := NoQ_RoomQ.RoomQ(ROOM_ID, ROOM_SECRET, ROOMQ_TICKET_ISSUER, ROOMQ_STATUS_API, ctx, true)
    if err := roomQ.Extend(ctx, 60); err != nil {
        ctx.IndentedJSON(http.StatusInternalServerError, Response{"error": err.Error()})
    } else {
        if serving, err := roomQ.GetServing(ctx); err != nil {
            ctx.IndentedJSON(http.StatusInternalServerError, Response{"error": err.Error()})
        } else {
            ctx.IndentedJSON(http.StatusOK, Response{"serving": serving})
        }
    }
}

func deleteTicket(ctx *gin.Context) {
    defer errorHandler(ctx)
    Log("Extend")
    roomQ := NoQ_RoomQ.RoomQ(ROOM_ID, ROOM_SECRET, ROOMQ_TICKET_ISSUER, ROOMQ_STATUS_API, ctx, true)
    if err := roomQ.DeleteServing(ctx); err != nil {
        ctx.IndentedJSON(http.StatusInternalServerError, Response{"error": err.Error()})
    } else {
        ctx.IndentedJSON(http.StatusOK, Response{})
    }
}

func errorHandler(ctx *gin.Context) {
    if err := recover(); err != nil {
        switch e := err.(type) {
        case *NoQ_RoomQ_Exception.InvalidApiKeyException:
            ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"error": e.Error()})
        case *NoQ_RoomQ_Exception.InvalidTokenException:
            ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"error": e.Error()})
        case *NoQ_RoomQ_Exception.NotServingException:
            ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": e.Error()})
        case *NoQ_RoomQ_Exception.QueueStoppedException:
            ctx.IndentedJSON(http.StatusGone, gin.H{"error": e.Error()})
        case *NoQ_RoomQ_Exception.ReachLimitException:
            ctx.IndentedJSON(http.StatusServiceUnavailable, gin.H{"error": e.Error()})
        default:
            Log(e)
            ctx.IndentedJSON(http.StatusAccepted, gin.H{})
        }
    }
}

func main() {
    router := gin.Default()
    router.GET("/api/roomq/get-ticket", getTicket)
    router.GET("/api/roomq/get-serving", getServing)
    router.GET("/api/roomq/extend-ticket", extendTicket)
    router.GET("/api/roomq/delete-ticket", deleteTicket)
    router.Run("0.0.0.0:3000")
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
| ------- | --------------- | ---------------------- |
| 1.x     | `github.com/redso/noq-roomq-go-sdk` | 1.17.1                    |
