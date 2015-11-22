package tank_utility
import (
	"encoding/json"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

//TODO(kendall): Refactor http client code (client creation + http requests)
//TODO(kendall): Investigate returning error from each method.
//TODO(kendall): Refactor URL building.
//TODO(kendall): Support reading a token file instead.
//TODO(kendall): Separate library from script to dump info.
//TODO(kendall): Replace Printf with logging and error/abort as appropriate.
//TODO(kendall): Reduce duplication.
//TODO(kendall): Add DeviceId to DeviceInfo struct.

type TankReading struct {
	Tank float64
	Temperature float32
	Time int64
	TimeIso string
}

type Device struct {
	Name string
	Address string
	Capacity int32
	LastReading TankReading
}

type DeviceInfo struct {
	Device Device
}

type DeviceList struct {
	Devices []string
}

type TokenResponse struct {
	Token string
}

func readCredentialsFile(credentials_file string) (string, string) {
	var credentials []byte
	var err error
	if credentials, err = ioutil.ReadFile(credentials_file); err != nil {
		fmt.Printf("Could not read credentials file: %s\n", err)
	}
	credential_parts := strings.SplitN(strings.TrimSpace(string(credentials)), ":", 2)
	return credential_parts[0], credential_parts[1]
}

func getHttpClient(insecure bool) (*http.Client) {
	var client *http.Client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if insecure {
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}
	return client
}

func callTankUtility(insecure bool, uri string, user string, password string, v interface{}) {
	var err error
	client := getHttpClient(insecure)
	req, req_err := http.NewRequest("GET", uri, nil)
	if req_err != nil {
		fmt.Printf("Request error: %s\n", req_err)
	}
	if user != "" && password != "" {
		req.SetBasicAuth(user, password)
	}
	resp, http_err := client.Do(req)
	if http_err != nil {
		fmt.Printf("Error: %s\n", http_err)
	} else {
		if json.NewDecoder(resp.Body).Decode(&v); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
}

func GetToken(credentials_file string, tank_utility_endpoint string, insecure bool) TokenResponse {
	user, password := readCredentialsFile(credentials_file)

	path := []string{tank_utility_endpoint, "getToken"}
	uri := strings.Join(path, "/")
	var token_response TokenResponse
	callTankUtility(insecure, uri, user, password, &token_response)
	return token_response
}

func GetDeviceList(token string, tank_utility_endpoint string, insecure bool) DeviceList {
	var devices_response DeviceList

	path := []string{tank_utility_endpoint, "devices"}
	uri := strings.Join(path, "/") + "?token=" + token
	callTankUtility(insecure, uri, "", "", &devices_response)
	return devices_response
}

func GetDeviceInfo(device string, token string, tank_utility_endpoint string, insecure bool) DeviceInfo {
	var device_response DeviceInfo

	path := []string{tank_utility_endpoint, "devices", device}
	uri := strings.Join(path, "/") + "?token=" + token
	callTankUtility(insecure, uri, "", "", &device_response)
	return device_response
}

func ReadTokenFromFile(token_file string) TokenResponse {
	var err error
	var token []byte
	var token_response TokenResponse
	if token, err = ioutil.ReadFile(token_file); err != nil {
		fmt.Printf("Could not read token file: %s\n", err)
	}
	if json.Unmarshal(token, &token_response); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	return token_response
}

func WriteTokenToFile(token_file string, token_response TokenResponse) {
	var err error
	var token []byte
	if token, err = json.Marshal(token_response); err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		ioutil.WriteFile(token_file, token, 0644)
	}
}
