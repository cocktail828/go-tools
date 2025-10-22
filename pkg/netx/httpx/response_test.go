package httpx

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestJSONStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type TestXMLStruct struct {
	XMLName xml.Name `xml:"user"`
	Name    string   `xml:"name"`
	Age     int      `xml:"age"`
}

func TestResponse_BindJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TestJSONStruct{Name: "test", Age: 20})
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	httpResp := &Response{Response: resp}
	var data TestJSONStruct
	assert.NoError(t, httpResp.BindJSON(&data))
	assert.Equal(t, "test", data.Name)
	assert.Equal(t, 20, data.Age)
}

func TestResponse_BindXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<user><name>test</name><age>20</age></user>`))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	httpResp := &Response{Response: resp}
	var data TestXMLStruct
	assert.NoError(t, httpResp.BindXML(&data))
	assert.Equal(t, "test", data.Name)
	assert.Equal(t, 20, data.Age)
}

func TestResponse_Bind_AutoContentType(t *testing.T) {
	jsonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TestJSONStruct{Name: "json", Age: 25})
	}))
	defer jsonServer.Close()

	xmlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<user><name>xml</name><age>30</age></user>`))
	}))
	defer xmlServer.Close()

	resp, _ := http.Get(jsonServer.URL)
	var jsonData TestJSONStruct
	assert.NoError(t, (&Response{Response: resp}).Bind(&jsonData))
	assert.Equal(t, "json", jsonData.Name)
	resp.Body.Close()

	resp, _ = http.Get(xmlServer.URL)
	var xmlData TestXMLStruct
	assert.NoError(t, (&Response{Response: resp}).Bind(&xmlData))
	assert.Equal(t, "xml", xmlData.Name)
	resp.Body.Close()
}

func TestResponse_Bind_Errors(t *testing.T) {
	noContentTypeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{}"))
	}))
	defer noContentTypeServer.Close()

	unsupportedTypeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("hello"))
	}))
	defer unsupportedTypeServer.Close()

	resp, _ := http.Get(noContentTypeServer.URL)
	var data TestJSONStruct
	assert.Error(t, (&Response{Response: resp}).Bind(&data))
	resp.Body.Close()

	resp, _ = http.Get(unsupportedTypeServer.URL)
	assert.Error(t, (&Response{Response: resp}).Bind(&data))
	resp.Body.Close()
}

func TestResponse_Bind_BodyCaching(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TestJSONStruct{Name: "cache", Age: 40})
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	httpResp := &Response{Response: resp}
	defer resp.Body.Close()

	var data1 TestJSONStruct
	assert.NoError(t, httpResp.Bind(&data1))

	var data2 TestJSONStruct
	assert.NoError(t, httpResp.Bind(&data2))

	assert.Equal(t, data1, data2)
}
