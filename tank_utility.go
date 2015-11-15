package main
import (
	"encoding/json"
	"crypto/tls"
	"flag"
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

var insecure = flag.Bool("insecure", true, "Whether to skip certificate checks.")
var credentials_file = flag.String("credentials_file", "", "Path to read username and pass from.")
var output_token_file = flag.String("output_token_file", "tank_utility.token", "Path to write the token to.")
var tank_utility_endpoint = flag.String("tank_utility_endpoint", "https://data.tankutility.com/api", "API endpoint for Tank Utility")

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
		fmt.Printf("Could not read credentials file: %s\n", credentials_file)
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
	fmt.Printf("Using %s\n", uri)
	if user != "" && password != "" {
		req.SetBasicAuth(user, password)
	}
	resp, http_err := client.Do(req)
	if http_err != nil {
		fmt.Printf("Error: %s\n", http_err)
	} else {
		var json_message []byte
		json_message, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		err = json.Unmarshal(json_message, &v)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
}

func GetToken(credentials_file string, tank_utility_endpoint string, insecure bool) string {
	user, password := readCredentialsFile(credentials_file)

	path := []string{tank_utility_endpoint, "getToken"}
	uri := strings.Join(path, "/")
	var token_response TokenResponse
	callTankUtility(insecure, uri, user, password, &token_response)
	return token_response.Token
}

func GetDeviceList(token string, tank_utility_endpoint string, insecure bool) []string {
	var devices_response DeviceList

	path := []string{tank_utility_endpoint, "devices"}
	uri := strings.Join(path, "/") + "?token=" + token
	callTankUtility(insecure, uri, "", "", &devices_response)
	return devices_response.Devices
}

func GetDeviceInfo(device string, token string, tank_utility_endpoint string, insecure bool) DeviceInfo {
	var device_response DeviceInfo

	path := []string{tank_utility_endpoint, "devices", device}
	uri := strings.Join(path, "/") + "?token=" + token
	callTankUtility(insecure, uri, "", "", &device_response)
	return device_response
}


func main() {
	flag.Parse()
	token := GetToken(*credentials_file, *tank_utility_endpoint, *insecure)
	device_list := GetDeviceList(token, *tank_utility_endpoint, *insecure)
	for i := 0; i < len(device_list); i++ {
		var device_info DeviceInfo
		device_info = GetDeviceInfo(device_list[0], token, *tank_utility_endpoint, *insecure)
		fmt.Printf("%#v\n", device_info)
	}
}
