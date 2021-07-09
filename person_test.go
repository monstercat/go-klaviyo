package klaviyo

import (
	"bytes"
	"encoding/json"
	"testing"
)

// A test person for our test cases. If you change this please make sure to update the tests!
func newTestPerson() Person {
	return Person{
		Object: Object{
			Id:     testPersonId,
			Object: "Person",
		},
		City:         "Vancouver",
		Consent:      []string{ConsentEmail, ConsentSMS},
		Country:      "Canada",
		Email:        "kitty@monstercat.com",
		FirstName:    "Kitty",
		LastName:     "Cat",
		Organization: "Monstercat",
		PhoneNumber:  "+1234567890",
		Region:       "British Columbia",
		Attributes:   map[string]interface{}{attrIsTest: true},
	}
}

func TestPerson_HasProfileIdentifier(t *testing.T) {
	p := newTestPerson()
	if p.HasProfileIdentifier() == false {
		t.Error("should have returned true")
	}
	p.PhoneNumber = ""
	if p.HasProfileIdentifier() == false {
		t.Error("should have returned true with no PhoneNumber but has Email")
	}
	p.Email = ""
	if p.HasProfileIdentifier() != false {
		t.Error("should have returned false with no PhoneNumber or Email")
	}
	p.PhoneNumber = "+1234567890"
	if p.HasProfileIdentifier() == false {
		t.Error("should have returned true with PhoneNumber but no Email")
	}
}

func TestPerson_GetMap(t *testing.T) {
	p := newTestPerson()
	m := p.GetMap()
	if m["$city"] != p.City {
		t.Error("Field City did not match map value.")
	}
	if m["$country"] != p.Country {
		t.Error("Field Country did not match map value.")
	}
	if m["$email"] != p.Email {
		t.Error("Field Email did not match map value.")
	}
	if m["$first_name"] != p.FirstName {
		t.Error("Field FirstName did not match map value.")
	}
	if m["$last_name"] != p.LastName {
		t.Error("Field LastName did not match map value.")
	}
	if m["$organization"] != p.Organization {
		t.Error("Field Organization did not match map value.")
	}
	if m["$phone_number"] != p.PhoneNumber {
		t.Error("Field PhoneNumber did not match map value.")
	}
	if m["$region"] != p.Region {
		t.Error("Field Region did not match map value.")
	}
	if m[attrIsTest] != p.Attributes[attrIsTest] {
		t.Error("Attribute IsTest did not match value.")
	}
	if arr, ok := m["$consent"].([]string); !ok {
		t.Error("$consent should be an array of strings")
	} else if len(arr) != len(p.Consent) {
		t.Errorf("Expected %d values for $consent.", len(p.Consent))
	}
}

func TestPerson_JSON(t *testing.T) {
	a := newTestPerson()
	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(&a); err != nil {
		t.Fatal(err)
	}
	var b Person
	if err := json.NewDecoder(buf).Decode(&b); err != nil {
		t.Fatal(err)
	}
	if a.City != b.City {
		t.Error("City did not match after encoding/decoding process")
	}
	if a.Country != b.Country {
		t.Error("Country did not match after encoding/decoding process")
	}
	if a.Email != b.Email {
		t.Error("Email did not match after encoding/decoding process")
	}
	if a.FirstName != b.FirstName {
		t.Error("FirstName did not match after encoding/decoding process")
	}
	if a.LastName != b.LastName {
		t.Error("LastName did not match after encoding/decoding process")
	}
	if a.Organization != b.Organization {
		t.Error("Organization did not match after encoding/decoding process")
	}
	if a.PhoneNumber != b.PhoneNumber {
		t.Error("PhoneNumber did not match after encoding/decoding process")
	}
	if a.Region != b.Region {
		t.Error("Region did not match after encoding/decoding process")
	}
	if len(a.Consent) != len(b.Consent) {
		t.Error("Consent length did not match")
	}
	if a.Attributes[attrIsTest] != b.Attributes[attrIsTest] {
		t.Error("Attribute did not match")
	}
}
