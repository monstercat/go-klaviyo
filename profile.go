package klaviyo

import (
	"encoding/json"
	"reflect"
	"strings"
)

type Person struct {
	Object

	// Below are special fields used by Klaviyo, they will be render in special UI for fancy-ness
	// they are identified by the $ prefix in their JSON.
	City         string   `json:"$city"`
	Consent      []string `json:"$consent"`
	Country      string   `json:"$country"`
	Email        string   `json:"$email"`
	FirstName    string   `json:"$first_name"`
	Image        string   `json:"$image"`
	LastName     string   `json:"$last_name"`
	Organization string   `json:"$organization"`
	PhoneNumber  string   `json:"$phone_number"`
	Region       string   `json:"$region"`
	Timezone     string   `json:"$timezone"`
	Title        string   `json:"$title"`
	Zip          string   `json:"$zip"`

	// Use these to have custom attributes tied to a user that can be used to create segments for lists.
	Attributes map[string]interface{}
}

// A profile identifier is an email or phone number. In the case of SMS they must have a phone number.
func (p *Person) HasProfileIdentifier() bool {
	return !(strings.TrimSpace(p.Email) == "" && strings.TrimSpace(p.PhoneNumber) == "")
}

func (p *Person) GetMap() map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range p.Attributes {
		m[k] = v
	}
	for k, v := range structToMap(p) {
		m[k] = v
	}
	return m
}

func (p *Person) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.GetMap())
}

func (p *Person) UnmarshalJSON(data []byte) error {
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, m); err != nil {
		return err
	}
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}

	// Remove the existing keys because they are part of the struct
	cleanKeys := []string{
		"id",
		"object",
		"$city",
		"$consent",
		"$country",
		"$email",
		"$first_name",
		"$image",
		"$last_name",
		"$organization",
		"$phone_number",
		"$region",
		"$timezone",
		"$title",
		"$zip",
	}
	for _, k := range cleanKeys {
		delete(m, k)
	}

	p.Attributes = m
	return nil
}

func structToMap(item interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	if item == nil {
		return res
	}
	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("json")
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = structToMap(field)
			} else {
				res[tag] = field
			}
		}
	}
	return res
}
