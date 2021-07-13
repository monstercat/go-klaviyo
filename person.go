package klaviyo

import (
	"encoding/json"
	"reflect"
	"strings"
)

type Attributes map[string]interface{}

func (a Attributes) ParseBool(key string) bool {
	val, ok := a[key]
	if !ok {
		return false
	}
	switch val.(type) {
	case string:
		if val.(string) == "true" || val.(string) == "1" {
			return true
		}
	case bool:
		return val.(bool)
	}
	return false
}

type Person struct {
	Object

	// Below are special fields used by Klaviyo, they will be render in special UI for fancy-ness
	// they are identified by the $ prefix in their JSON.
	// Please read here for more: https://help.klaviyo.com/hc/en-us/articles/115005084927-Template-Tags-and-Variable-Syntax#klaviyo-special-properties18
	//
	// Any extra attributes appear in the same flat structure but we store them in Attributes below.
	CustomId     string   `json:"$id"`
	Address1     string   `json:"$address1"`
	Address2     string   `json:"$address2"`
	City         string   `json:"$city"`
	Consent      []string `json:"$consent"`
	Country      string   `json:"$country"`
	Email        string   `json:"$email"`
	FirstName    string   `json:"$first_name"`
	Image        string   `json:"$image"`
	LastName     string   `json:"$last_name"`
	Latitude     string   `json:"$latitude"`
	Longitude    string   `json:"$longitude"`
	Organization string   `json:"$organization"`
	PhoneNumber  string   `json:"$phone_number"`
	Region       string   `json:"$region"`
	Source       string   `json:"$source"`
	Timezone     string   `json:"$timezone"`
	Title        string   `json:"$title"`
	Zip          string   `json:"$zip"`

	// Use these to have custom attributes tied to a user that can be used to create segments for lists.
	Attributes Attributes
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
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	type Person2 Person
	var p2 Person2
	if err := json.Unmarshal(data, &p2); err != nil {
		return err
	}

	// Remove keys natively supported by klaviyo
	delete(m, "id")
	delete(m, "object")
	for k, _ := range m {
		if len(k) <= 0 {
			continue
		}
		if k[0] != '$' {
			continue
		}
		delete(m, k)
	}

	*p = Person(p2)
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
