package main

import (
	"fmt"
	"testing"
)

func TestShouldGenerateCode(t *testing.T) {
	n := 6
	code := generateCode(n)

	if codeLength := len(code); codeLength != n {
		t.Errorf("The length of the generated code should be equal to %d. Code length %d given", n, codeLength)
	}
}

func TestShouldMarshalLinkResponse(t *testing.T) {
	l := newLink("abcd", "http://golang.org")
	expectedJson := fmt.Sprintf("{\"code\":\"%s\",\"short_path\":\"/code/%s\",\"url\":\"%s\",\"created_at\":\"%s\"}", l.Code, l.Code, l.Url, l.CreatedAt)
	json, err := l.MarshalJSON()
	if err != nil {
		t.Errorf("Error %s was not expected at the marshalling time", err)
	}

	if expectedJson != string(json) {
		t.Errorf("Wrong response given at the marshalling time:\n %s\n Expected: %s\n", string(json), expectedJson)
	}
}
