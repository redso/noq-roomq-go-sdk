package NoQ_RoomQ

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	NoQ_RoomQ_Exception "github.com/redso/noq-roomq-go-sdk/Exception"
	NoQ_RoomQ_Utils "github.com/redso/noq-roomq-go-sdk/Utils"
)

type roomQ struct {
	clientID       string
	jwtSecret      string
	ticketIssuer   string
	debug          bool
	tokenName      string
	token          string
	statusEndpoint string
}

func RoomQ(clientID string, jwtSecret string, ticketIssuer string, statusEndpoint string, ctx *gin.Context, debug bool) roomQ {
	rQ := roomQ{
		clientID:       clientID,
		jwtSecret:      jwtSecret,
		ticketIssuer:   ticketIssuer,
		debug:          debug,
		statusEndpoint: statusEndpoint,
		tokenName:      fmt.Sprintf("be_roomq_t_%s", clientID),
	}
	rQ.token = rQ.GetToken(ctx)
	return rQ
}

func (rQ roomQ) GetToken(ctx *gin.Context) string {
	if token, ok := ctx.GetQuery("noq_t"); ok {
		return token
	}
	if token, err := ctx.Cookie(rQ.tokenName); err == nil {
		return token
	}
	return ""
}

func (rQ *roomQ) Validate(ctx *gin.Context, returnURL, sessionID string) validationResult {
	token := rQ.token
	currentURL := ""
	if ctx.Request.TLS != nil {
		currentURL = fmt.Sprintf("https://%s%s", ctx.Request.Host, ctx.Request.URL.Path)
	} else {
		currentURL = fmt.Sprintf("http://%s%s", ctx.Request.Host, ctx.Request.URL.Path)
	}
	needGenerateJWT := false
	needRedirect := false
	if len(token) < 1 {
		needGenerateJWT = true
		needRedirect = true
		rQ.debugPrint("no jwt")
	} else {
		rQ.debugPrint("current jwt " + token)
		if data, ok := NoQ_RoomQ_Utils.JwtDecode(token, rQ.jwtSecret); ok {
			if len(sessionID) > 0 && data.Get("session_id").String() != sessionID {
				needGenerateJWT = true
				needRedirect = true
				rQ.debugPrint("session id not match")
			} else if data.HasKey("deadline") && data.Get("deadline").Int() < time.Now().UTC().UnixMilli() {
				needRedirect = true
				rQ.debugPrint("deadline exceed")
			} else if data.Get("type").String() == "queue" {
				needRedirect = true
				rQ.debugPrint("in queue")
			} else if data.Get("type").String() == "self-sign" {
				needRedirect = true
				rQ.debugPrint("self sign token")
			}
		} else {
			needGenerateJWT = true
			needRedirect = true
			rQ.debugPrint("Failed to decode jwt")
			rQ.debugPrint("invalid secret")
		}
	}
	if needGenerateJWT {
		token = rQ.generateJWT(sessionID)
		rQ.debugPrint("generating new jwt token")
		rQ.token = token
	}
	ctx.SetCookie(rQ.tokenName, rQ.token, int(time.Now().UnixMilli())+(12*60*60), "/", "", false, false)
	if needRedirect {
		if len(returnURL) > 0 {
			return rQ.redirectToTicketIssuer(token, returnURL)
		} else {
			return rQ.redirectToTicketIssuer(token, currentURL)
		}
	} else {
		return rQ.enter(currentURL)
	}
}

func (rQ *roomQ) Extend(ctx *gin.Context, duration int) error {
	if backend, err := rQ.GetBackend(); err == nil {
		httpClient := NoQ_RoomQ_Utils.HttpClient(fmt.Sprintf("https://%s", backend))
		response := httpClient.Post(fmt.Sprintf("/queue/%s", rQ.clientID), map[string]interface{}{
			"action":                  "beep",
			"client_id":               rQ.clientID,
			"id":                      rQ.token,
			"extend_serving_duration": duration * 60,
		})
		rQ.debugPrint(response.StatusCode)
		if response.StatusCode == http.StatusUnauthorized {
			rQ.debugPrint(&NoQ_RoomQ_Exception.InvalidApiKeyException{})
			return &NoQ_RoomQ_Exception.InvalidApiKeyException{}
		} else if response.StatusCode == http.StatusNotFound {
			rQ.debugPrint(&NoQ_RoomQ_Exception.NotServingException{})
			return &NoQ_RoomQ_Exception.NotServingException{}
		} else {
			token := response.Get("id").String()
			rQ.token = token
			ctx.SetCookie(rQ.tokenName, rQ.token, int(time.Now().UnixMilli())+(12*60*60), "/", "", false, false)
			return nil
		}
	} else {
		rQ.debugPrint(err)
		return err.(error)
	}
}

func (rQ *roomQ) GetServing(ctx *gin.Context) (int64, error) {
	if backend, err := rQ.GetBackend(); err == nil {
		httpClient := NoQ_RoomQ_Utils.HttpClient(fmt.Sprintf("https://%s", backend))
		response := httpClient.Get(fmt.Sprintf("/rooms/%s/servings/%s", rQ.clientID, rQ.token))
		rQ.debugPrint(response.Raw)
		if response.StatusCode == http.StatusUnauthorized {
			rQ.debugPrint(&NoQ_RoomQ_Exception.InvalidApiKeyException{})
			return 0, &NoQ_RoomQ_Exception.InvalidApiKeyException{}
		} else if response.StatusCode == http.StatusNotFound {
			rQ.debugPrint(&NoQ_RoomQ_Exception.NotServingException{})
			return 0, &NoQ_RoomQ_Exception.NotServingException{}
		} else {
			return int64(response.Get("deadline").Float()) / 1000, nil
		}
	} else {
		rQ.debugPrint(err)
		return 0, err.(error)
	}
}

func (rQ *roomQ) DeleteServing(ctx *gin.Context) error {
	if backend, err := rQ.GetBackend(); err == nil {
		httpClient := NoQ_RoomQ_Utils.HttpClient(fmt.Sprintf("https://%s/queue", backend))
		response := httpClient.Post(fmt.Sprintf("/%s", rQ.clientID), map[string]interface{}{
			"action":    "delete_serving",
			"client_id": rQ.clientID,
			"id":        rQ.token,
		})
		rQ.debugPrint(response.Raw)
		if response.StatusCode == http.StatusUnauthorized {
			rQ.debugPrint(&NoQ_RoomQ_Exception.InvalidApiKeyException{})
			return &NoQ_RoomQ_Exception.InvalidApiKeyException{}
		} else if response.StatusCode == http.StatusNotFound {
			rQ.debugPrint(&NoQ_RoomQ_Exception.NotServingException{})
			return &NoQ_RoomQ_Exception.NotServingException{}
		} else {
			if payload, ok := NoQ_RoomQ_Utils.JwtDecode(rQ.token, rQ.jwtSecret); ok {
				token := rQ.generateJWT(payload.Get("session_id").String())
				rQ.token = token
				ctx.SetCookie(rQ.tokenName, rQ.token, int(time.Now().UnixMilli())+(12*60*60), "/", "", false, false)
				return nil
			} else {
				return errors.New("Failed to decode jwt")
			}
		}
	} else {
		rQ.debugPrint(err)
		return err.(error)
	}
}

func (rQ roomQ) enter(currentURL string) validationResult {
	urlWithoutToken := removeNoQToken(currentURL)
	// redirect if url contain token
	if urlWithoutToken != currentURL {
		return ValidationResult(urlWithoutToken)
	}
	return ValidationResult("")
}

func (rQ roomQ) redirectToTicketIssuer(token, currentURL string) validationResult {
	urlWithoutToken := removeNoQToken(currentURL)
	params := url.Values{}
	params.Add("noq_t", token)
	params.Add("noq_c", rQ.clientID)
	params.Add("noq_r", urlWithoutToken)
	rQ.debugPrint(params)
	if base, err := url.Parse(rQ.ticketIssuer); err == nil {
		base.RawQuery = params.Encode()
		return ValidationResult(base.String())
	} else {
		rQ.debugPrint("Failed to redirect to ticket issuer")
		panic("Failed to redirect to ticket issuer")
	}
}

func (rQ roomQ) generateJWT(sessionID string) string {
	_sessionID := ""
	if len(sessionID) > 0 {
		_sessionID = sessionID
	} else if _uuid, err := uuid.NewRandom(); err == nil {
		_sessionID = _uuid.String()
	}
	claims := NoQ_RoomQ_Utils.JwtClaims{
		RoomID:    rQ.clientID,
		SessionID: _sessionID,
		Type:      "self-sign",
	}
	return NoQ_RoomQ_Utils.JwtEncode(claims, rQ.jwtSecret)
}

func (rQ roomQ) debugPrint(message interface{}) {
	if rQ.debug {
		log.Println(fmt.Sprintf("[RoomQ] %s", message))
	}
}

func removeNoQToken(currentURL string) string {
	url := regexp.MustCompile("(?i)([&]*)(noq_t=[^&]*)").ReplaceAllString(currentURL, "")
	url = regexp.MustCompile("(?i)(\\?&)").ReplaceAllString(url, "?")
	url = regexp.MustCompile("(?i)(\\?$)").ReplaceAllString(url, "")
	return url
}

func (rQ roomQ) GetBackend() (string, interface{}) {
	client := NoQ_RoomQ_Utils.HttpClient(rQ.statusEndpoint)
	resp := client.Get(fmt.Sprintf("/%s", rQ.clientID))
	if resp.Get("state").String() == "stopped" {
		return "", &NoQ_RoomQ_Exception.QueueStoppedException{}
	}
	return resp.Get("backend").String(), nil
}
