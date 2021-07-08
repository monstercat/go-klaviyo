// Klaviyo uses Person & Person interchangeably through their API documentation, we will use just Person
// https://apidocs.klaviyo.com/reference/api-overview

package klaviyo

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Consent string
type Reason string

const (
	ConsentEmail  Consent = "email"
	ConsentWeb    Consent = "web"
	ConsentSms    Consent = "sms"
	ConsentDirect Consent = "directmail"
	ConsentMobile Consent = "mobile"

	// They have multiple endpoints unfortunately.
	Endpoint   = "https://a.klaviyo.com/api"
	EndpointV1 = "https://a.klaviyo.com/api/v1"
	EndpointV2 = "https://a.klaviyo.com/api/v2"
)

var (
	ErrNoProfileIdentifier = errors.New("there is no unique profile identifier, must have email or phone number")
	ErrBadResponse         = errors.New("bad API response")
)

type APIError struct {
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

type Object struct {
	Id     string `json:"id"`
	Object string `json:"object"` // e.g. person, $list
}

type Client struct {
	// Sometimes called "token"
	PublicKey string

	// Sometimes seen as "api_key"
	PrivateKey string

	// The amount of time an HTTP API call should run for before it times out.
	DefaultTimeout time.Duration
}

func (c *Client) req(method string, url *url.URL, out interface{}) error {
	url.Query().Add("api_key", c.PrivateKey)
	req, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	client := http.Client{Timeout: c.DefaultTimeout}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	// All of Klaviyo's calls should return 200 otherwise it's an error.
	// See more here: https://apidocs.klaviyo.com/reference/api-overview#errors
	if res.StatusCode != http.StatusOK {
		var err APIError
		if err := json.NewDecoder(res.Body).Decode(&err); err != nil {
			return ErrBadResponse
		}
		return &err
	}
	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}
	return nil
}

// https://apidocs.klaviyo.com/reference/track-identify#identify
// GET https://a.klaviyo.com/api/identify
func (c *Client) Identify(person *Person) error {
	if !person.HasProfileIdentifier() {
		return ErrNoProfileIdentifier
	}

	payload := struct {
		Token      string      `json:"token"`
		Properties interface{} `json:"properties"`
	}{
		Token:      c.PublicKey,
		Properties: person.GetMap(),
	}
	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(&payload); err != nil {
		return err
	}
	u := newEndpoint(Endpoint, "identify")
	u.Query().Add("data", base64.URLEncoding.EncodeToString(buf.Bytes()))
	return c.req(http.MethodGet, u, nil)
}

// https://apidocs.klaviyo.com/reference/profiles#get-profile
// GET https://a.klaviyo.com/api/v1/person/person_id
func (c *Client) GetPerson(personId string) (*Person, error) {
	var p *Person
	err := c.req(http.MethodGet, newEndpoint(EndpointV1, fmt.Sprintf("person/%s", personId)), p)
	return p, err
}

// https://apidocs.klaviyo.com/reference/profiles#update-profile
// PUT https://a.klaviyo.com/api/v1/person/person_id
// Only works to update a persons attributes after they have been identified.
func (c *Client) UpdatePerson(person *Person) error {
	u := newEndpoint(EndpointV1, fmt.Sprintf("person/%s", person.Id))
	for k, v := range person.GetMap() {
		u.Query().Add(k, fmt.Sprintf("%v", v))
	}
	return c.req(http.MethodPut, u, person)
}

type ListPerson struct {
	Id          string `json:"id"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Created     string `json:"created"`
}

// https://apidocs.klaviyo.com/reference/lists-segments#list-membership
// GET https://a.klaviyo.com/api/v2/list/list_id/members
func (c *Client) InList(listId string, emails []string, phoneNumbers []string, pushTokens []string) ([]ListPerson, error) {
	u := newEndpoint(EndpointV2, fmt.Sprintf("list/%s/members", listId))
	u.Query().Add("emails", strings.Join(emails, ","))
	u.Query().Add("phone_numbers", strings.Join(phoneNumbers, ","))
	u.Query().Add("push_tokens", strings.Join(pushTokens, ","))
	var res []ListPerson
	err := c.req(http.MethodGet, u, &res)
	return res, err
}

func newEndpoint(endpoint, path string) *url.URL {
	u, err := url.Parse(endpoint)
	if err != nil {
		panic(err) // This should always work because endpoint should be typed correctly in this SDK!
	}
	u.Path = path
	return u
}
