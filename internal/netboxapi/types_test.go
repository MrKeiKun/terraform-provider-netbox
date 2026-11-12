package netboxapi

import (
	"errors"
	"net/http"
	"testing"
)

func TestTypeHelpers(t *testing.T) {
	if RefID(nil) != nil || RefID(&Ref{ID: 3}) != int64(3) {
		t.Error("RefID")
	}
	if got := IDs([]Ref{{ID: 1}, {ID: 2}}); len(got) != 2 || got[1] != 2 {
		t.Errorf("IDs = %v", got)
	}
	if len(IDs(nil)) != 0 {
		t.Error("IDs of nothing")
	}
	if ChoiceValue(nil) != "" || ChoiceValue(&Choice{Value: "active"}) != "active" {
		t.Error("ChoiceValue")
	}
	if IntChoiceValue(nil) != nil || IntChoiceValue(&IntChoice{Value: 10}) != int64(10) {
		t.Error("IntChoiceValue")
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{Method: "GET", URL: "http://x/api/a/", StatusCode: http.StatusNotFound, Body: "gone"}
	if err.Error() != "netbox: GET http://x/api/a/ returned 404: gone" {
		t.Errorf("message = %q", err.Error())
	}
	if !IsNotFound(err) || IsNotFound(errors.New("x")) || IsNotFound(&APIError{StatusCode: 500}) {
		t.Error("IsNotFound")
	}
}
