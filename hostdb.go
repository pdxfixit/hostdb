package hostdb

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// ErrorResponse is used for any requests which result in an error
type ErrorResponse struct {
	Code    int
	Message string
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Message)
}

// GenericError is used for all sorts of things
type GenericError struct {
	Error string `json:"error"`
}

// GetCatalogQuantityResponse is used for catalog requests with the count query parameter
type GetCatalogQuantityResponse struct {
	Count     int            `json:"count"`
	QueryTime string         `json:"query_time"`
	Catalog   map[string]int `json:"catalog"`
}

// GetCatalogResponse is used for catalog requests
type GetCatalogResponse struct {
	Count     int      `json:"count"`
	QueryTime string   `json:"query_time"`
	Catalog   []string `json:"catalog"`
}

// GetHealthResponse is used for responding to health check requests
type GetHealthResponse struct {
	App string `json:"app"`
	DB  string `json:"db"`
}

// GetRecordsResponse is used for responding to record requests
type GetRecordsResponse struct {
	Count     int               `json:"count"`
	QueryTime string            `json:"query_time"`
	Records   map[string]Record `json:"records"`
}

// GetStatsResponse is used for responding to stats requests
type GetStatsResponse struct {
	Hostname           string            `json:"hostname"`
	TotalRecords       int               `json:"total_records"`
	NewestRecord       string            `json:"newest_record"`
	OldestRecord       string            `json:"oldest_record"`
	LastSeenCollectors map[string]string `json:"lastseen_collectors"`
}

// GetVersionResponse is used for version requests
type GetVersionResponse struct {
	App ServerVersion  `json:"app"`
	DB  MariadbVersion `json:"db"`
}

// GlobalConfig contains all the different configurations needed for operation
type GlobalConfig struct {
	Hostdb  ServerConfig  `json:"hostdb" mapstructure:"hostdb"`
	Mariadb MariadbConfig `json:"mariadb" mapstructure:"mariadb"`
	API     struct {
		Version string      `json:"version" mapstructure:"version"`
		V0      APIv0Config `json:"v0" mapstructure:"v0"`
	} `json:"api" mapstructure:"api"`
}

// PostRecordsResponse is used to respond to POST requests
type PostRecordsResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// PutRecordResponse is used when responding to single-record PUT requests
type PutRecordResponse struct {
	ID string `json:"id"`
	OK bool   `json:"ok"`
}

// Record is the standard HostDB schema
type Record struct {
	ID        string                 `json:"id,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Hostname  string                 `json:"hostname,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Committer string                 `json:"committer,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Data      json.RawMessage        `json:"data,omitempty"`
	Hash      string                 `json:"hash,omitempty"`
}

// Send will PUT a single record into HostDB
func (r Record) Send(uniqueIdentifier string) (err error) {

	// ensure we have a unique identifier, which is used when viewing logs
	if uniqueIdentifier == "" {
		uniqueIdentifier = r.Type
	}

	// post data to HostDB
	responseBytes, err := httpRequest(
		"PUT",
		fmt.Sprintf("/records/%s", r.ID),
		r,
		nil,
	)
	if err != nil {
		log.Println(fmt.Sprintf("%s", string(responseBytes)))
		return err
	}

	// unmarshal the response into a struct
	var response PutRecordResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return err
	}

	// if something's not OK, return an error
	if response.OK != true {
		return fmt.Errorf("failed to save record %s", response.ID)
	}

	log.Println(fmt.Sprintf("1 %s record sent to HostDB", r.Type))

	return nil

}

// RecordSet is a collection of similar records
type RecordSet struct {
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
	Committer string                 `json:"committer,omitempty"`
	Records   []Record               `json:"records"`
}

// Save will write a JSON file to disk, with what would be submitted to HostDB.
// Will attempt to create directory if it doesn't exist.
// Defaults to /sample-data/<type>.json
func (rs RecordSet) Save(filePath string) (err error) {

	if filePath == "" {
		filePath = fmt.Sprintf("%s/%s.json", "/sample-data", rs.Type)
	} else if filepath.Ext(filePath) != ".json" {
		return errors.New("provided file path must end in .json")
	}

	// convert the struct into bytes
	requestBytes, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	// let the user know we're starting
	log.Println(fmt.Sprintf("saving %d records into %s", len(rs.Records), filePath))

	// ensure path exists for output
	if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			log.Fatal(err)
		}
	}

	// output to file
	if err := ioutil.WriteFile(filePath, requestBytes, 0644); err != nil {
		log.Fatal(err)
	}

	// let the user know we're done
	log.Println(fmt.Sprintf("saved %d records into %s", len(rs.Records), filePath))

	return nil

}

// Send will post the RecordSet to HostDB, if credentials are present.
// The input variable uniqueQueryString is used when viewing traffic logs --
// it has no impact on function
func (rs RecordSet) Send(uniqueQueryString string) (err error) {

	// ensure we have a unique identifier, which is used when viewing logs
	if uniqueQueryString == "" {
		uniqueQueryString = fmt.Sprintf("?type=%s", rs.Type)
	}

	// let the user know we're starting
	log.Println(fmt.Sprintf("sending %d %v record(s) to HostDB", len(rs.Records), rs.Type))

	// post data to HostDB
	responseBytes, err := httpRequest(
		"POST",
		fmt.Sprintf("/records/?%s", uniqueQueryString),
		rs,
		nil,
	)
	if err != nil {
		log.Println(fmt.Sprintf("%s", string(responseBytes)))
		return err
	}

	// unmarshal the response into a struct
	var response PostRecordsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return err
	}

	// if something's not OK, return an error
	if response.OK != true {
		return errors.New(response.Error)
	}

	// let the user know we're done
	log.Println(fmt.Sprintf("sent %d %v record(s) to HostDB", len(rs.Records), rs.Type))

	return nil

}

// ServerConfig contains the configuration parameters for hostdb-server
type ServerConfig struct {
	Host               string `json:"host" mapstructure:"host"`
	Port               int    `json:"port" mapstructure:"port"`
	Pass               string `json:"pass" mapstructure:"pass"`
	URL                string `json:"url" mapstructure:"url"`
	Debug              bool   `json:"debug" mapstructure:"debug"`
	NewRelicAppName    string `json:"newrelic_appname" mapstructure:"newrelic_appname"`
	NewRelicLicenseKey string `json:"newrelic_license" mapstructure:"newrelic_license"`
}

// ServerVersion contains the relevant aspects of the current binary version
type ServerVersion struct {
	Version    string `json:"version"`
	APIVersion string `json:"api_version"`
	Commit     string `json:"commit"`
	Date       string `json:"date"`
	BuildURL   string `json:"build_url"`
	GoVersion  string `json:"go_version"`
}

func httpRequest(method string, path string, requestBody interface{}, header map[string]string) (responseBytes []byte, err error) {

	// hostdbURL
	hostdbURL, found := os.LookupEnv("HOSTDB_URL")
	if !found || hostdbURL == "" {
		hostdbURL = "https://hostdb.pdxfixit.com/v0"
	} else {
		if _, err := url.ParseRequestURI(hostdbURL); err != nil {
			log.Fatal("HOSTDB_URL is invalid")
		}
	}
	hostdbURL = hostdbURL + path

	// hostdbUser
	hostdbUser, found := os.LookupEnv("HOSTDB_USER")
	if !found || hostdbUser == "" {
		hostdbUser = "writer"
	}

	// hostdbPass
	hostdbPass := os.Getenv("HOSTDB_PASS")

	// if no password, return
	if hostdbPass == "" {
		log.Println("no password, canceling request")
		return []byte(`{"OK":true,"error":"no password, canceling request"}`), nil
	}

	//log.Println(fmt.Sprintf("using ***%s : ***%s @ %s", hostdbUser[len(hostdbUser)-3:], hostdbPass[len(hostdbPass)-3:], hostdbURL))

	// encode creds for basic auth
	if len(header) < 1 {
		header = make(map[string]string)
	}
	header["Authorization"] = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(
		"%s:%s", hostdbUser, hostdbPass,
	))))

	// convert the struct into bytes
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// Client
	client := &http.Client{}
	req, err := http.NewRequest(method, hostdbURL, bytes.NewReader(requestBytes))
	if err != nil {
		log.Fatal(err)
	}

	// INSECURE
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// headers
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}

	var res *http.Response
	res, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	responseBytes, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		return responseBytes, errors.New(res.Status)
	}

	return responseBytes, nil

}
