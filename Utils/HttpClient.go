package NoQ_RoomQ_Utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type httpClient struct {
	baseURL string
}

func HttpClient(baseURL string) httpClient {
	return httpClient{baseURL: baseURL}
}

type request struct {
	path    string
	payload io.Reader
}

func (client httpClient) makeRequest(method string, req request) JSON {
	log.Println("[HttpClient]: " + client.baseURL + req.path)
	reqHandler := &http.Client{}
	if request, err := http.NewRequest(method, client.baseURL+req.path, req.payload); err == nil {
		request.Header.Set("Content-Type", "application/json; charset=utf-8")
		if resp, err := reqHandler.Do(request); err == nil {
			defer resp.Body.Close()
			if body, err := io.ReadAll(resp.Body); err == nil {
				var data JSON
				data.Raw = string(body)
				data.StatusCode = resp.StatusCode
				if err := json.Unmarshal(body, &data.Val); err == nil {
					return data
				} else {
					panic(err)
				}
			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}
}

func (client httpClient) Get(path string) JSON {
	return client.makeRequest(http.MethodGet, request{path: path})
}

func (client httpClient) Put(path string, payload map[string]interface{}) JSON {
	if param, err := json.Marshal(payload); err == nil {
		return client.makeRequest(http.MethodPut, request{path: path, payload: bytes.NewBuffer(param)})
	} else {
		panic(err)
	}
}

func (client httpClient) Post(path string, payload map[string]interface{}) JSON {
	if param, err := json.Marshal(payload); err == nil {
		return client.makeRequest(http.MethodPost, request{path: path, payload: bytes.NewBuffer(param)})
	} else {
		panic(err)
	}
}

func (client httpClient) Delete(path string) JSON {
	return client.makeRequest(http.MethodDelete, request{path: path})
}
