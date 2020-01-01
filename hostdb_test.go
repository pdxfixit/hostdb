package hostdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var TestRecord = Record{
	ID:        "",
	Type:      "",
	Hostname:  "",
	IP:        "",
	Timestamp: "",
	Committer: "",
	Context:   map[string]interface{}{},
	Data:      json.RawMessage{},
	Hash:      "",
}

var TestRecordSet = RecordSet{
	Type:      "test",
	Timestamp: "2003-04-05 06:07:08",
	Context: map[string]interface{}{
		"test": true,
	},
	Committer: "Test Monkey",
	Records:   []Record{TestRecord},
}

func TestErrorResponse_Error(t *testing.T) {

	err := ErrorResponse{
		Code:    500,
		Message: "testing",
	}

	assert.Equal(t, fmt.Sprintf("%v: %v", err.Code, err.Message), err.Error(), "error message")

}

func TestRecord_Send(t *testing.T) {

	// fake http server for posting records
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, fmt.Sprintf("{\"ok\":true,\"id\":\"%s\"}", TestRecord.ID))
		if err != nil {
			t.Error(err.Error())
		}
	}))
	defer testServer.Close()

	if err := os.Setenv("HOSTDB_URL", testServer.URL); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("HOSTDB_USER", "user"); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("HOSTDB_PASS", "pass"); err != nil {
		t.Fatal(err.Error())
	}

	if err := TestRecord.Send("test"); err != nil {
		t.Fatal(err.Error())
	}

}

func TestRecordSet_Save(t *testing.T) {

	filename := "sample-data/test.json"

	if err := TestRecordSet.Save(filename); err != nil {
		t.Fatal(err.Error())
	}

	assert.FileExists(t, filename, "saving a sample data file")

	// verify that the sample data file contents match what we provided

	// convert the struct into bytes
	requestBytes, err := json.Marshal(TestRecordSet)
	if err != nil {
		t.Fatal(err.Error())
	}

	// open the sample data file
	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(err.Error())
	}

	// retrieve the contents of the written file
	fileBytes, err := ioutil.ReadAll(file)
	if err := file.Close(); err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, requestBytes, fileBytes, "comparing request to file")
}

func TestRecordSet_Send(t *testing.T) {

	// fake http server for posting records
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "{\"ok\":true,\"error\":\"1 record(s) processed\"}")
		if err != nil {
			t.Error(err.Error())
		}
	}))
	defer testServer.Close()

	if err := os.Setenv("HOSTDB_URL", testServer.URL); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("HOSTDB_USER", "user"); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("HOSTDB_PASS", "pass"); err != nil {
		t.Fatal(err.Error())
	}

	// post the records
	if err := TestRecordSet.Send("test"); err != nil {
		t.Errorf("%v", err)
	}

}

func TestRecordSet_Send_WithNoCreds(t *testing.T) {

	if err := os.Setenv("HOSTDB_URL", ""); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("HOSTDB_USER", ""); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("HOSTDB_PASS", ""); err != nil {
		t.Fatal(err.Error())
	}

	if err := TestRecordSet.Send("test"); err != nil {
		t.Errorf("%v", err)
	}

}
