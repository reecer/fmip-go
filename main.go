package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

const AUTH_URL = "https://fmipmobile.icloud.com/fmipservice/device/"

var AUTH_ERROR = errors.New("Invalid username/password pair")
var IMPL_ERROR = errors.New("Unknown Error: The API could have changed. Please report this.")

type Location struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
	Type      string  `json:"positionType"`
}

type Device struct {
	Location Location `json:"location"`
	Name     string   `json:"name"`
	Id       string   `json:"id"`
}

type FMIPReturn struct {
	Content []Device `json:"content"`
}

type FMIPClient struct {
	User     string
	Password string
	Host     string
	Scope    string
	BaseUrl  string
}

// Create a new FMIPClient
func New(uname string, pword string) FMIPClient {
	return FMIPClient{
		User:     uname,
		Password: pword,
		BaseUrl:  AUTH_URL + uname + "/",
	}
}

// Private request helper (provides Auth header)
func (c FMIPClient) req(url string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.BaseUrl+url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// fmt.Println(c.BaseUrl+url, data)
	req.SetBasicAuth(c.User, c.Password)
	client := &http.Client{}
	return client.Do(req)
}

// Retreive Host with login
func (c *FMIPClient) Login() error {
	resp, err := c.req("initClient", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode > 400 {
		return AUTH_ERROR
	}

	// Pretty sure Scope isn't even necessary
	c.Host = resp.Header.Get("X-Apple-MMe-Host")
	c.Scope = resp.Header.Get("X-Apple-MMe-Scope")
	c.BaseUrl = "https://" + c.Host + "/fmipservice/device/" + c.Scope + "/"
	return nil
}

// Get devices
func (c FMIPClient) Devices() ([]Device, error) {

	resp, err := c.req("initClient", nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, IMPL_ERROR
	}

	var ret FMIPReturn
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret.Content, nil
}

// Send a message to a device
func (c FMIPClient) Message(device Device, title string, msg string, sound bool) error {
	body := struct {
		Device   string `json:"device"`
		Subject  string `json:"subject"`
		Text     string `json:"text"`
		UserText bool   `json:"userText"`
		Sound    bool   `json:"sound"`
	}{device.Id, title, msg, true, sound}

	data, _ := json.Marshal(body)

	resp, err := c.req("sendMessage", []byte(data))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return IMPL_ERROR
	}
	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Expected uname and pword")
		return
	}

	uname := os.Args[1]
	pword := os.Args[2]

	c := New(uname, pword)
	if err := c.Login(); err != nil {
		fmt.Println(err)
		return
	}

	devices, err := c.Devices()
	if err != nil {
		fmt.Println(err)
		return
	}

	d := devices[0]

	fmt.Println(d.Location)
	c.Message(d, "123", "ABC", false)
}
