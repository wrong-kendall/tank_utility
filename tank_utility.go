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

var insecure = flag.Bool("insecure", true, "Whether to skip certificate checks.")
var credentials_file = flag.String("credentials_file", "", "Path to read username and pass from.")
var output_token_file = flag.String("output_token_file", "tank_utility.token", "Path to write the token to.")
var tank_utility_endpoint = flag.String("tank_utility_endpoint", "https://data.tankutility.com/api", "API endpoint for Tank Utility")

func GetToken(credentials_file string, tank_utility_endpoint string, insecure bool) string {
	var credentials []byte
	var err error
	if credentials, err = ioutil.ReadFile(credentials_file); err != nil {
		fmt.Printf("Could not read credentials file: %s\n", credentials_file)
	}
	credential_parts := strings.SplitN(strings.TrimSpace(string(credentials)), ":", 2)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client *http.Client
	if insecure {
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	path := []string{tank_utility_endpoint, "getToken"}
	uri := strings.Join(path, "/")
	req, req_err := http.NewRequest("GET", uri, nil)
	if req_err != nil {
		fmt.Printf("Request error: %s\n", req_err)
	}
	fmt.Printf("Using %s\n", uri)
	req.SetBasicAuth(credential_parts[0], credential_parts[1])
	resp, http_err := client.Do(req)
	if http_err != nil {
		fmt.Printf("Error: %s\n", http_err)
	} else {
		type Message struct {
			Token string
		}
		var token_response Message
		var json_message []byte
		json_message, err = ioutil.ReadAll(resp.Body)
		fmt.Printf("Message: %s\n", json_message)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		err = json.Unmarshal(json_message, &token_response)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		return token_response.Token
	}
	return ""
}

func GetDeviceList(token string, tank_utility_endpoint string, insecure bool) []string {
	var err error
	var client *http.Client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if insecure {
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	path := []string{tank_utility_endpoint, "devices"}
	uri := strings.Join(path, "/") + "?token=" + token
	fmt.Printf("Using %s\n", uri)
	resp, http_err := client.Get(uri)
	if http_err != nil {
		fmt.Printf("Error: %s\n", http_err)
	} else {
		type Message struct {
			Devices []string
		}
		var devices_response Message
		var json_message []byte
		json_message, err = ioutil.ReadAll(resp.Body)
		fmt.Printf("Message: %s\n", json_message)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		err = json.Unmarshal(json_message, &devices_response)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		return devices_response.Devices
	}
	return make([]string, 0)
}


func main() {
	flag.Parse()
	token := GetToken(*credentials_file, *tank_utility_endpoint, *insecure)
	fmt.Printf("Token: %s\n", token)
	fmt.Printf("Devices: %s", GetDeviceList(token, *tank_utility_endpoint, *insecure))
}
