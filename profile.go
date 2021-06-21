package klaviyo

import "errors"

type Person struct {
	Object

	// Below are special fields used by Klaviyo, they will be render in special UI for fancy-ness
	// they are identified by the $ prefix in their JSON.
	City         string `json:"$city"`
	Consent      []string `json:"$consent"`
	Country      string `json:"$country"`
	Email        string `json:"$email"`
	FirstName    string `json:"$first_name"`
	Image        string `json:"$image"`
	LastName     string `json:"$last_name"`
	Organization string `json:"$organization"`
	PhoneNumber  string `json:"$phone_number"`
	Region       string `json:"$region"`
	Timezone     string `json:"$timezone"`
	Title        string `json:"$title"`
	Zip          string `json:"$zip"`

	// Use these to have custom attributes tied to a user that can be used to create segments for lists.
	Attributes   map[string]interface{}
}

func (p *Person) MarshalJSON() ([]byte, error) {
	// TODO convert Person to a map[string]interface{} that overrides the attributes map
	return nil, errors.New("not implemented")
}

func (p *Person) UnmarshalJSON(data []byte) error {
	// TODO unmarshal data into profile and all others go into attributes
	return errors.New("not implemented")
}
