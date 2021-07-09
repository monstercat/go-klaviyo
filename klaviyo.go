// Klaviyo uses profile & person interchangeably through their API documentation, we will use just Person
// https://apidocs.klaviyo.com/reference/api-overview

package klaviyo

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Consent string

const (
	ConsentEmail  Consent = "email"
	ConsentWeb    Consent = "web"
	ConsentSms    Consent = "sms"
	ConsentDirect Consent = "directmail"
	ConsentMobile Consent = "mobile"

	// Use these instead of the MIME library because this is what is specified in their documentation.
	ContentHTML = "text/html"
	ContentJSON = "application/json"

	// They have multiple endpoints unfortunately.
	Endpoint   = "https://a.klaviyo.com/api"
	EndpointV1 = "https://a.klaviyo.com/api/v1"
	EndpointV2 = "https://a.klaviyo.com/api/v2"
)

var (
	ErrNoProfileIdentifier = errors.New("there is no unique profile identifier, must have email or phone number")
	ErrFailed              = errors.New("request successful, call failed")
	ErrInvalidOutArg       = errors.New("out arg provided does not match datatype of response")
)

func newEndpoint(endpoint, path string) *url.URL {
	u, err := url.Parse(endpoint)
	if err != nil {
		panic(err) // This should always work because endpoint should be typed correctly in this SDK!
	}
	u.Path = path
	return u
}

type BadResponseError struct {
	Body      []byte
	JSONError error
}

func (e *BadResponseError) Error() string {
	return "bad response"
}

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

func (c *Client) req(method, accept string, url *url.URL, out interface{}) error {
	url.Query().Add("api_key", c.PrivateKey)
	req, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", accept)
	client := http.Client{Timeout: c.DefaultTimeout}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	contentType := res.Header.Get("Content-Type")
	var data []byte
	if buf, err := io.ReadAll(res.Body); err != nil {
		return err
	} else {
		data = buf
	}
	// All of Klaviyo's calls should return 200 otherwise it's an error.
	// See more here: https://apidocs.klaviyo.com/reference/api-overview#errors
	if res.StatusCode != http.StatusOK {
		var err APIError
		if contentType != ContentJSON {
			err.Message = string(data)
		} else {
			if jsonErr := json.NewDecoder(bytes.NewBuffer(data)).Decode(&err); jsonErr != nil {
				return &BadResponseError{
					Body:      data,
					JSONError: jsonErr,
				}
			}
		}
		return &err
	}
	if out != nil {
		switch contentType {
		case ContentJSON:
			return json.NewDecoder(bytes.NewBuffer(data)).Decode(out)
		case ContentHTML:
			k, ok := out.(*string)
			if !ok {
				return ErrInvalidOutArg
			}
			*k = string(data)
		}
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
	var res string
	if err := c.req(http.MethodGet, ContentHTML, u, &res); err != nil {
		return err
	}
	if res != "1" {
		return ErrFailed
	}
	return nil
}

// https://apidocs.klaviyo.com/reference/profiles#get-profile
// GET https://a.klaviyo.com/api/v1/person/person_id
func (c *Client) GetPerson(personId string) (*Person, error) {
	var p Person
	err := c.req(http.MethodGet, ContentJSON, newEndpoint(EndpointV1, fmt.Sprintf("person/%s", personId)), &p)
	return &p, err
}

// https://apidocs.klaviyo.com/reference/profiles#update-profile
// PUT https://a.klaviyo.com/api/v1/person/person_id
// Only works to update a persons attributes after they have been identified.
func (c *Client) UpdatePerson(person *Person) error {
	u := newEndpoint(EndpointV1, fmt.Sprintf("person/%s", person.Id))
	for k, v := range person.GetMap() {
		u.Query().Add(k, fmt.Sprintf("%v", v))
	}
	return c.req(http.MethodPut, ContentJSON, u, person)
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
	err := c.req(http.MethodGet, ContentJSON, u, &res)
	return res, err
}
