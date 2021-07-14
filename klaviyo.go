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
	"path"
	"strings"
	"time"
)

const (
	ConsentEmail  = "email"
	ConsentWeb    = "web"
	ConsentSMS    = "sms"
	ConsentDirect = "directmail"
	ConsentMobile = "mobile"

	// Use these instead of the MIME library because this is what is specified in their documentation.
	ContentNone     = ""
	ContentHTML     = "text/html"
	ContentHTMLUTF8 = "text/html; charset=utf-8"
	ContentJSON     = "application/json"

	// They have multiple endpoints unfortunately.
	Endpoint   = "https://a.klaviyo.com/api"
	EndpointV1 = "https://a.klaviyo.com/api/v1"
	EndpointV2 = "https://a.klaviyo.com/api/v2"
)

var (
	ErrNoPublicKey         = errors.New("missing public key")
	ErrNoPrivateKey        = errors.New("missing private key")
	ErrNoProfileIdentifier = errors.New("there is no unique profile identifier, must have email or phone number")
	ErrFailed              = errors.New("request successful, call failed")
	ErrInvalidOutArg       = errors.New("out arg provided does not match datatype of response")
)

func newEndpoint(endpoint, uri string) *url.URL {
	u, err := url.Parse(endpoint)
	if err != nil {
		panic(err) // This should always work because endpoint should be typed correctly in this SDK!
	}
	u.Path = path.Join(u.Path, uri)
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
	// Use this to store the raw error response if the response is not parseable.
	Raw string

	// Klaviyo's documentation details the usage of "message", but returns "detail" in some instances.
	Detail  string `json:"detail"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	} else if e.Detail != "" {
		return e.Detail
	}
	return e.Raw
}

// All objects in Klaviyo use this basic structure to identify what kind of object it is and how to identify it.
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

func (c *Client) doReq(r *http.Request, out interface{}) error {
	// We are adding the private key on all requests because it is easier to do.
	if c.PrivateKey == "" {
		return ErrNoPrivateKey
	}
	values := r.URL.Query()
	values.Add("api_key", c.PrivateKey)
	r.URL.RawQuery = values.Encode()

	client := http.Client{Timeout: c.DefaultTimeout}
	res, err := client.Do(r)
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
		err.Raw = string(data)
		return &err
	}
	if out != nil {
		switch contentType {
		case ContentJSON:
			return json.NewDecoder(bytes.NewBuffer(data)).Decode(out)
		case ContentHTML, ContentHTMLUTF8:
			k, ok := out.(*string)
			if !ok {
				return ErrInvalidOutArg
			}
			*k = string(data)
		}
	}
	return nil
}

func (c *Client) send(method, accept string, url *url.URL, out interface{}) error {
	req, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", accept)
	return c.doReq(req, out)
}

func (c *Client) sendJSON(method, accept string, url *url.URL, in interface{}, out interface{}) error {
	xs, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, url.String(), bytes.NewReader(xs))
	if err != nil {
		return err
	}
	if accept != ContentNone {
		req.Header.Add("Accept", accept)
	}
	req.Header.Add("Content-Type", ContentJSON)
	return c.doReq(req, out)
}

// https://apidocs.klaviyo.com/reference/track-identify#identify
// GET https://a.klaviyo.com/api/identify
// TODO Update Identify to use POST method version as GET is outdated
func (c *Client) Identify(person *Person) error {
	return c.IdentifySafe(person, false)
}

// Use this if you do not want to send values that are not set. This is great for when you want to update a Person
// without first fetching their information. This will happen if you only have thier email and no Klaviyo Id to utilize.
func (c *Client) IdentifySafe(person *Person, omit bool) error {
	if c.PublicKey == "" {
		return ErrNoPublicKey
	}
	if !person.HasProfileIdentifier() {
		return ErrNoProfileIdentifier
	}

	props := person.GetMap()
	if omit {
		trimEmptyValues(props)
	}

	payload := struct {
		Token      string      `json:"token"`
		Properties interface{} `json:"properties"`
	}{
		Token:      c.PublicKey,
		Properties: props,
	}
	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(&payload); err != nil {
		return err
	}
	u := newEndpoint(Endpoint, "identify")
	values := u.Query()
	values.Add("data", base64.StdEncoding.EncodeToString(buf.Bytes()))
	u.RawQuery = values.Encode()
	var res string
	if err := c.send(http.MethodGet, ContentHTML, u, &res); err != nil {
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
	err := c.send(http.MethodGet, ContentJSON, newEndpoint(EndpointV1, fmt.Sprintf("person/%s", personId)), &p)
	return &p, err
}

// https://apidocs.klaviyo.com/reference/profiles#update-profile
// PUT https://a.klaviyo.com/api/v1/person/person_id
// Only works to update a persons attributes after they have been identified.
func (c *Client) UpdatePerson(person *Person) error {
	u := newEndpoint(EndpointV1, fmt.Sprintf("person/%s", person.Id))
	values := u.Query()
	for k, v := range person.GetMap() {
		values.Add(k, fmt.Sprintf("%v", v))
	}
	u.RawQuery = values.Encode()
	return c.send(http.MethodPut, ContentJSON, u, person)
}

// https://apidocs.klaviyo.com/reference/lists-segments#subscribe
// POST https://a.klaviyo.com/api/v2/list/list_id/subscribe
func (c *Client) Subscribe(listId string, emails, phoneNumbers []string) ([]ListPerson, error) {
	u := newEndpoint(EndpointV2, fmt.Sprintf("list/%s/subscribe", listId))
	var res []ListPerson
	type payload struct {
		Profiles []map[string]interface{} `json:"profiles"`
	}
	p := payload{
		Profiles: []map[string]interface{}{},
	}
	for _, email := range emails {
		p.Profiles = append(p.Profiles, map[string]interface{}{
			"email": email,
		})
	}
	for _, num := range phoneNumbers {
		p.Profiles = append(p.Profiles, map[string]interface{}{
			"phone_number": num,
			"sms_consent":  true,
		})
	}
	err := c.sendJSON(http.MethodPost, ContentJSON, u, &p, &res)
	return res, err
}

// https://apidocs.klaviyo.com/reference/lists-segments#unsubscribe
// DELETE https://a.klaviyo.com/api/v2/list/list_id/subscribe
func (c *Client) Unsubscribe(listId string, emails, phoneNumbers, pushTokens []string) error {
	u := newEndpoint(EndpointV2, fmt.Sprintf("list/%s/subscribe", listId))
	toc := map[string][]string{
		"emails":        emails,
		"phone_numbers": phoneNumbers,
		"push_tokens":   pushTokens,
	}
	m := map[string][]string{}
	for k, arr := range toc {
		if len(arr) > 0 {
			m[k] = make([]string, 0)
		}
		for _, x := range arr {
			m[k] = append(m[k], x)
		}
	}
	return c.sendJSON(http.MethodDelete, ContentNone, u, m, nil)
}

type ListPerson struct {
	Id          string `json:"id"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Created     string `json:"created"`
}

// https://apidocs.klaviyo.com/reference/lists-segments#list-membership
// GET https://a.klaviyo.com/api/v2/list/list_id/members
func (c *Client) InList(listId string, emails, phoneNumbers, pushTokens []string) ([]ListPerson, error) {
	u := newEndpoint(EndpointV2, fmt.Sprintf("list/%s/members", listId))
	if len(emails) == 0 && len(phoneNumbers) == 0 && len(pushTokens) == 0 {
		return nil, nil
	}
	values := u.Query()
	if len(emails) > 0 {
		values.Add("emails", strings.Join(emails, ","))
	}
	if len(phoneNumbers) > 0 {
		values.Add("phone_numbers", strings.Join(phoneNumbers, ","))
	}
	if len(pushTokens) > 0 {
		values.Add("push_tokens", strings.Join(pushTokens, ","))
	}
	u.RawQuery = values.Encode()
	var res []ListPerson
	err := c.send(http.MethodGet, ContentJSON, u, &res)
	return res, err
}

func trimEmptyValues(m map[string]interface{}) map[string]interface{} {
	for key, val := range m {
		var kill bool
		switch val.(type) {
		case nil:
			kill = true
		case string:
			if val.(string) == "" {
				kill = true
			}
		}
		if kill {
			delete(m, key)
		}
	}
	return m
}