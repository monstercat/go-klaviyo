package klaviyo

import (
	"os"
	"testing"
	"time"
)

var (
	testPersonId  = os.Getenv("KlaviyoTestPersonId")
	testListId    = os.Getenv("KlaviyoTestListId")
)

const (
	attrIsTest    = "IsTest"
	attrLikesGold = "LikesGold"
)

func newTestClient() *Client {
	return &Client{
		PublicKey:      os.Getenv("KlaviyoPublicKey"),
		PrivateKey:     os.Getenv("KlaviyoPrivateKey"),
		DefaultTimeout: time.Second,
	}
}

func TestClient_Identify(t *testing.T) {
	client := newTestClient()
	p := newTestPerson()
	err := client.Identify(&p)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetPerson(t *testing.T) {
	client := newTestClient()
	p, err := client.GetPerson(testPersonId)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("Returned person was nil")
	}
}

func TestClient_UpdatePerson(t *testing.T) {
	client := newTestClient()
	p, err := client.GetPerson(testPersonId)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("Returned person was nil")
	}
	t.Log("attr likes gold", p.Attributes[attrLikesGold])
	likesGold := p.Attributes.ParseBool(attrLikesGold)
	t.Log("parsed likes gold", likesGold)

	likesGold = !likesGold
	p.Attributes[attrLikesGold] = likesGold
	err = client.UpdatePerson(p)
	if err != nil {
		t.Fatal(err)
	}

	// Verify update went through
	b, err := client.GetPerson(p.Id)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("b attr likes gold", b.Attributes[attrLikesGold])
	if _, ok := b.Attributes[attrLikesGold]; !ok {
		t.Fatalf("Did not find attribute %s", attrLikesGold)
	} else if b.Attributes.ParseBool(attrLikesGold) != likesGold {
		t.Fatalf("Attribute did not match for %s", attrLikesGold)
	}
}

func TestClient_InList(t *testing.T) {
	client := newTestClient()
	p := newTestPerson()

	// This checks to make sure the test user is in the test list
	xs, err := client.InList(testListId, []string{p.Email}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(xs) != 1 {
		t.Fatalf("Expected 1 ListPerson in array")
	}
	if xs[0].Email != p.Email {
		t.Fatalf("Returned ListPerson.Email does not match input")
	}

	// This checks that a real user is not in the test list
	xs, err = client.InList(testListId, []string{"dev@monstercat.com"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(xs) != 0 {
		t.Fatalf("User should not appear in the test list!")
	}
}

// This test expects that your list is using single opt-in settings. Double opt-in will not return any results.
func TestClient_Subscribe(t *testing.T) {
	email := "dev@monstercat.com"
	client := newTestClient()
	// TODO get list information on double opt-in status to adapt test checks
	res, err := client.Subscribe(testListId, []string{email}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 {
		t.Fatal("Expected 1 result back from Subscribe call, please make sure that you are using single opt-in")
	} else if res[0].Email != email {
		t.Fatalf("Result email did not match input email")
	}
}

func TestClient_Unsubscribe(t *testing.T) {
	email := "dev@monstercat.com"
	client := newTestClient()
	err := client.Unsubscribe(testListId, []string{email}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
}
