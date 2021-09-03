package klaviyo

import (
	"regexp"
	"strconv"
)

var (
	frontBackQuotesRegexp = regexp.MustCompile("^\"|\"$")
)

// KFloat implements the UnmarshalJSON interface to do special processing for Klaviyo. In certain instances (such as
// when a number field is empty), klaviyo will return the value as a string. Otherwise, it will return the value as a
// number.
type KFloat float64

func (f *KFloat) UnmarshalJSON(b []byte) error {
	// Strip the quotes from the front and back of the JSON string. This allows string values of floats such as
	// "123.345" to pass properly. However, quotes in the middle will not be removed.
	s := frontBackQuotesRegexp.ReplaceAll(b, nil)

	// Parse to a float. This function will fail on empty string.
	v, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		return err
	}
	*f = KFloat(v)
	return nil
}

// KInt implements the UnmarshalJSON interface to do special processing for Klaviyo. In certain instances (such as
// when a number field is empty), klaviyo will return the value as a string. Otherwise, it will return the value as a
// number.
type KInt int

func (i *KInt) UnmarshalJSON(b []byte) error {
	// Strip the quotes from the front and back of the JSON string. This allows string values of floats such as
	// "12345" to pass properly. However, quotes in the middle will not be removed.
	s := frontBackQuotesRegexp.ReplaceAll(b, nil)

	v, err := strconv.Atoi(string(s))
	if err != nil {
		return err
	}
	*i = KInt(v)
	return nil 
}
