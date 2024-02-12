package dto

import (
	"encoding/json"
	"testing"
)

type StructWithOptionalTitle struct {
	Title       *Title `json:"title,omitempty"`
	Description string `json:"description"`
}

func TestTitle_UnmarshalJSON(t *testing.T) {

	for _, testCase := range []struct {
		encodedStruct []byte
		decodedStruct StructWithOptionalTitle
	}{
		{
			[]byte("{\"title\":\"foo bar\",\"description\":\"baz\"}"),
			StructWithOptionalTitle{Title: NewTitle("foo bar"), Description: "baz"},
		},
		{
			[]byte("{\"title\":0,\"description\":\"baz\"}"),
			StructWithOptionalTitle{Title: NewTitle(""), Description: "baz"},
		},
		{
			[]byte("{\"description\":\"baz\"}"),
			StructWithOptionalTitle{Title: nil, Description: "baz"},
		},
	} {
		var decodedStruct StructWithOptionalTitle
		err := json.Unmarshal(testCase.encodedStruct, &decodedStruct)
		if err != nil {
			t.Errorf("unmarshal failed: %+v", err)
		}

		if !(decodedStruct.Title == nil && testCase.decodedStruct.Title == nil) &&
			*decodedStruct.Title != *testCase.decodedStruct.Title {
			t.Errorf("unexpected unmarshal result. expected %+v, got %+v", testCase.decodedStruct.Title, decodedStruct.Title)
		}
	}
}

func TestTitle_MarshalJSON(t *testing.T) {
	type StructWithTitle struct {
		Title       *Title `json:"title,omitempty"`
		Description string `json:"description"`
	}

	for _, testCase := range []struct {
		decodedStruct StructWithTitle
		encodedStruct []byte
	}{
		{
			StructWithTitle{Title: NewTitle("foo bar"), Description: "baz"},
			[]byte("{\"title\":\"foo bar\",\"description\":\"baz\"}"),
		},
		{
			StructWithTitle{Title: NewTitle(""), Description: "baz"},
			[]byte("{\"title\":0,\"description\":\"baz\"}"),
		},
		{
			StructWithTitle{Title: nil, Description: "baz"},
			[]byte("{\"description\":\"baz\"}"),
		},
	} {
		encodedStruct, err := json.Marshal(testCase.decodedStruct)
		if err != nil {
			t.Errorf("unmarshal failed: %+v", err)
		}

		if string(encodedStruct) != string(testCase.encodedStruct) {
			t.Errorf("unexpected marshal result. expected %s, got %s", testCase.encodedStruct, encodedStruct)
		}
	}
}
