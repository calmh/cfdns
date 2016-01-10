package cfdns

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultAPIBase = "https://api.cloudflare.com/client/v4"
)

// A DNSRecord contains information about a single DNS Record.
type DNSRecord struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	Content   string      `json:"content"`
	ID        string      `json:"id,omitempty"`
	TTL       int         `json:"ttl,omitempty"`
	Proxiable bool        `json:"proxiable"`
	Proxied   bool        `json:"proxied"`
	Locked    bool        `json:"locked"`
	ZoneID    string      `json:"zone_id,omitempty"`
	ZoneName  string      `json:"zone_name,omitempty"`
	Created   time.Time   `json:"created_on,omitempty"`
	Modified  time.Time   `json:"modified_on,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func (r DNSRecord) String() string {
	return r.Name + " " + r.Type + " " + r.Content
}

type cfDNSResult struct {
	Result []DNSRecord
}

type Zone struct {
	ID                  string    `json:"id,omitempty"`
	Name                string    `json:"name"`
	DevelopmentMode     int       `json:"development_mode,omitempty"`
	OriginalRegistrar   string    `json:"original_registrar,omitempty"`
	OriginalDNSHost     string    `json:"original_dnshost,omitempty"`
	OriginalNameServers []string  `json:"original_name_servers,omitempty"`
	Created             time.Time `json:"created_on,omitempty"`
	Modified            time.Time `json:"modified_on,omitempty"`
	NameServers         []string  `json:"name_servers,omitempty"`
	Status              string    `json:"status,omitempty"`
	Paused              bool      `json:"paused,omitempty"`
	Type                string    `json:"type,omitempty"`
	Checked             time.Time `json:"checked_on,omitempty"`
}

type cfZoneResult struct {
	Result []Zone
}

// A Client is used to communicate with the Cloudflare service.
type Client struct {
	apiBase   string
	authEmail string
	authKey   string
}

// NewClient returns a new, initialized Client. The email and API key are used to authenticate against Cloudflare.
func NewClient(authEmail, authKey string) *Client {
	return &Client{
		apiBase:   defaultAPIBase,
		authEmail: authEmail,
		authKey:   authKey,
	}
}

// ListZones returns a list of all zones accessable by the client, or an error.
func (c *Client) ListZones() ([]Zone, error) {
	resp, err := c.doRequest("GET", "/zones?per_page=1000", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res cfZoneResult
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	return res.Result, nil
}

// ListDNSRecords returns a list of all DNS records under the given zone, or an error. The zone ID can be found in the results of ListZones.
func (c *Client) ListDNSRecords(zoneID string) ([]DNSRecord, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/zones/%s/dns_records?per_page=1000", zoneID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res cfDNSResult
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	return res.Result, nil
}

// CreateDNSRecord creates a new DNS record in the given zone, with the specified name, type and content.
func (c *Client) CreateDNSRecord(zoneID, name, rectype, content string) error {
	body, _ := jsonReader(DNSRecord{
		Name:    name,
		Type:    rectype,
		Content: content,
	})

	resp, err := c.doRequest("POST", fmt.Sprintf("/zones/%s/dns_records", zoneID), body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}

// UpdateDNSRecord updates an existing record and returns nil, or an error. The record should be as recieved from ListDNSRecords; specifically, the fields ID and ZoneID must be filled out.
func (c *Client) UpdateDNSRecord(rec DNSRecord) error {
	body, err := jsonReader(rec)
	if err != nil {
		return err
	}

	resp, err := c.doRequest("PUT", fmt.Sprintf("/zones/%s/dns_records/%s", rec.ZoneID, rec.ID), body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}

// UpdateDNSRecord deletes an existing record and returns nil, or an error. The record should be as recieved from ListDNSRecords; specifically, the fields ID and ZoneID must be filled out.
func (c *Client) DeleteDNSRecord(rec DNSRecord) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/zones/%s/dns_records/%s", rec.ZoneID, rec.ID), nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}

func (c *Client) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.apiBase, url), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Email", c.authEmail)
	req.Header.Set("X-Auth-Key", c.authKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func jsonReader(obj interface{}) (io.Reader, error) {
	bs, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(bs), nil
}
