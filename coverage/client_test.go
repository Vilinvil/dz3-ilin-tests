package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

const (
	errorBadRequest      = "unknown bad request error: request invalid"
	errorLimitBelowZero  = "limit must be > 0"
	errorOffsetBelowZero = "offset must be > 0"
	errorBadAccessToken  = "bad AccessToken"
	errorServer          = "SearchServer fatal error. Body: {\"Error\":\"Internal server error\"}"
	errorTimeout         = "timeout for limit=2&offset=0&order_by=1&order_field=Id&query="
	errorWrongUrl        = "unknown error Get \"wrongUrl?limit=2&offset=0&order_by=1&order_field=Id&query=\": unsupported protocol scheme \"\""
	errorNotFound        = "users not found"
)

type testCaseClient struct {
	Request     SearchRequest
	Result      *SearchResponse
	ResponseErr error
}

func TestClient(t *testing.T) {
	cases := []testCaseClient{
		{
			Request: SearchRequest{Limit: 1,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    1},

			Result: &SearchResponse{Users: []User{{ID: 0,
				Name:   "Boyd Wolf",
				Age:    22,
				About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
				Gender: "male"}},
				NextPage: true},

			ResponseErr: nil,
		},
		{
			Request: SearchRequest{Limit: -1,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    1},

			Result:      nil,
			ResponseErr: fmt.Errorf(errorLimitBelowZero),
		},
		{
			Request: SearchRequest{Limit: 50,
				Offset:     0,
				Query:      "This text don't find",
				OrderField: "Id",
				OrderBy:    1},

			Result:      nil,
			ResponseErr: fmt.Errorf(errorNotFound),
		},
		{
			Request: SearchRequest{Limit: 2,
				Offset:     -1,
				Query:      "",
				OrderField: "Id",
				OrderBy:    1},

			Result:      nil,
			ResponseErr: fmt.Errorf(errorOffsetBelowZero),
		},
		{
			Request: SearchRequest{Limit: 5,
				Offset:     0,
				Query:      "",
				OrderField: "BadField",
				OrderBy:    1},
			Result:      nil,
			ResponseErr: fmt.Errorf(errorBadRequest),
		},
		{
			Request: SearchRequest{Limit: 5,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    987654},
			Result:      nil,
			ResponseErr: fmt.Errorf(errorBadRequest),
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	testClient := SearchClient{AccessToken: "2a54a886a8bbcc309ae4ffa75241cd6d", URL: ts.URL}

	for caseNum, item := range cases {
		result, err := testClient.FindUsers(item.Request)
		if !reflect.DeepEqual(err, item.ResponseErr) {
			t.Errorf("[%d] got unexpected error: %#v, expected: %#v", caseNum, err, item.ResponseErr)
		}

		if item.ResponseErr != nil && err == nil {
			t.Errorf("[%d] got: %v expected error: %#v", caseNum, err, item.ResponseErr)
		}

		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d] wrong result, got: %#v, expected: %#v,", caseNum, result, item.Result)
		}
	}
	ts.Close()
}

func TestClientSpecificError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	testClient := SearchClient{AccessToken: "wrongToken", URL: ts.URL}
	req := SearchRequest{Limit: 1,
		Offset:     0,
		Query:      "",
		OrderField: "Id",
		OrderBy:    1}

	result, err := testClient.FindUsers(req)
	if err.Error() != errorBadAccessToken {
		t.Errorf("Error is: %v. Result is: %v", err, result)
	}
	testClient.AccessToken = "2a54a886a8bbcc309ae4ffa75241cd6d"

	client.Timeout = time.Microsecond
	result, err = testClient.FindUsers(req)
	if err.Error() != errorTimeout {
		t.Errorf("Error is: %v. Result is: %v", err, result)
	}
	client.Timeout = time.Second

	PatchDataSet = "wrongDataSet"
	result, err = testClient.FindUsers(req)
	if err.Error() != errorServer {
		t.Errorf("Error is: %v. Result is: %v", err, result)
	}
	PatchDataSet = "data_set.xml"

	testClient.URL = "wrongUrl"
	result, err = testClient.FindUsers(req)
	if err.Error() != errorWrongUrl {
		t.Errorf("Error is: %v. Result is: %v", err, result)
	}
	ts.Close()
}

/*
	go test -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html

*/
// тут писать код тестов
