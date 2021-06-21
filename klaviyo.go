// Klaviyo uses Person & Person interchangeably through their API documentation, we will use just Person

package klaviyo

import "errors"

type Consent string
type Reason string

const (
	ConsentEmail  Consent = "email"
	ConsentWeb    Consent = "web"
	ConsentSms    Consent = "sms"
	ConsentDirect Consent = "directmail"
	ConsentMobile Consent = "mobile"

	ReasonBounced      Reason = "Bounced"
	ReasonUnsubscribed Reason = "Unsubscribed"

	EndpointV1 = "https://a.klaviyo.com/api/v1"
	EndpointV2 = "https://a.klaviyo.com/api/v2"
)

var (
	ErrNil = errors.New("not implemented")
)

type Object struct {
	Id           string `json:"id"`
	Object       string `json:"object"` // person,
}

type Client struct {
	APIKey string
}

func (c *Client) GetPerson(personId string) (*Person, error) {
	// GET v1/person/%s?api_key=%s
	return nil, ErrNil
}

func (c *Client) UpdatePerson(person *Person) error {
	// PUT v1/person/%s?api_key=%s
	return ErrNil
}
