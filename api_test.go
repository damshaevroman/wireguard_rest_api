package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"wireguard_api/usecases"
)

//   THIS E2E testing for test need on ubutnu 24
//   Start application complie go build main.go start  -  sudo ./main
//   Go to direcroty with file api_test.go
//   Start test with command go test -v api_test.go

const baseURL = "https://127.0.0.1:8888"
const bearerToken = "12345"

type errResponse struct {
	Result string `json:"result"`
}

type clientGetResponse struct {
	Result []usecases.ClientResponse `json:"result"`
}

type interfaceGetResponse struct {
	Result []usecases.ServerInterfaces `json:"result"`
}

var serverPrivate string

var insecureClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func sendRequest(t *testing.T, method, path string, body any) (int, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL+path, reader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := insecureClient.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	return resp.StatusCode, bodyBytes
}

func TestCreateInterface(t *testing.T) {
	body := map[string]any{
		"ifname":   "testapi",
		"ip":       "192.168.32.10/24",
		"endpoint": "10.19.44.251",
		"port":     10010,
	}
	status, rawData := sendRequest(t, "POST", "/interface/new", body)

	if status == 200 {
		var data usecases.ServerInterfaces
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())
		}
		serverPrivate = data.Private
	} else {
		var data errResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())
		}
	}
}

func TestStopInterface(t *testing.T) {
	body := map[string]any{"ifname": "testapi"}
	status, rawData := sendRequest(t, "POST", "/interface/stop", body)
	if status != 200 {
		var data errResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())

		}
	}
}

func TestStartInterface(t *testing.T) {
	body := map[string]any{"ifname": "testapi"}
	status, rawData := sendRequest(t, "POST", "/interface/start", body)
	if status != 200 {
		var data errResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())

		}
	}
}

func TestCreateClient(t *testing.T) {
	body := map[string]any{
		"ifname":   "testapi",
		"ip":       "",
		"alloweip": "",
		"dns":      "8.8.8.8",
	}
	status, rawData := sendRequest(t, "POST", "/clients/new", body)

	if status != 200 {
		var data errResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())

		}
	}
}

func TestGetClients(t *testing.T) {
	status, rawData := sendRequest(t, "GET", "/clients/getall", nil)
	if status != 200 {
		var data errResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())

		}
	} else {
		var data clientGetResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Get clients failed: %s", err.Error())
		}
		if len(data.Result) == 0 {
			t.Fatalf("No clients found")
		}
		for _, client := range data.Result {

			body := map[string]string{"public": client.Public}

			status, rawData := sendRequest(t, "DELETE", "/clients", body)
			if status != 200 {
				var data errResponse
				err := json.Unmarshal(rawData, &data)
				if err != nil {
					t.Fatalf("Create interface failed: %s", err.Error())

				}
			}
		}
	}
}

func TestGetClientStatus(t *testing.T) {
	status, rawData := sendRequest(t, "GET", "/clients/status", nil)
	if status != 200 {
		var data errResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())

		}
	}
}

func TestGetInterfaces(t *testing.T) {
	status, rawData := sendRequest(t, "GET", "/interface/all", nil)
	if status != 200 {
		var data errResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())

		}
	} else {
		var data interfaceGetResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Get clients failed: %s", err.Error())
		}
		for _, client := range data.Result {
			body := map[string]string{"private": client.Private, "ifname": client.Ifname}
			status, rawData := sendRequest(t, "DELETE", "/interface", body)
			if status != 200 {
				var data errResponse
				err := json.Unmarshal(rawData, &data)
				if err != nil {
					t.Fatalf("Create interface failed: %s", err.Error())

				}
			}
		}
	}
}

func TestGetArchiveInterfaces(t *testing.T) {
	status, rawData := sendRequest(t, "GET", "/interface/archive", nil)
	if status != 200 {
		var data errResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())
		}
	}
}

func TestGetVersion(t *testing.T) {
	status, rawData := sendRequest(t, "GET", "/version", nil)
	if status != 200 {
		var data errResponse
		err := json.Unmarshal(rawData, &data)
		if err != nil {
			t.Fatalf("Create interface failed: %s", err.Error())
		}
	}
}
