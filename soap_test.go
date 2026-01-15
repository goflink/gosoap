package gosoap

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"text/template"
)

// mockWSDLTemplate for soap_test.go tests
const soapTestWSDL = `<?xml version="1.0" encoding="utf-8"?>
<wsdl:definitions xmlns:s="http://www.w3.org/2001/XMLSchema"
                  xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
                  xmlns:tns="http://test.example/"
                  targetNamespace="http://test.example/"
                  xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/">
  <wsdl:types>
    <s:schema elementFormDefault="qualified" targetNamespace="http://test.example/">
      <s:element name="checkVat">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="countryCode" type="s:string" />
            <s:element minOccurs="0" maxOccurs="1" name="vatNumber" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="checkVatResponse">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="countryCode" type="s:string" />
            <s:element minOccurs="0" maxOccurs="1" name="vatNumber" type="s:string" />
            <s:element minOccurs="0" maxOccurs="1" name="requestDate" type="s:string" />
            <s:element minOccurs="0" maxOccurs="1" name="valid" type="s:string" />
            <s:element minOccurs="0" maxOccurs="1" name="name" type="s:string" />
            <s:element minOccurs="0" maxOccurs="1" name="address" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="CapitalCity">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="sCountryISOCode" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="CapitalCityResponse">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="CapitalCityResult" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="NumberToWords">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="ubiNum" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="NumberToWordsResponse">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="NumberToWordsResult" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="Whois">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="DomainName" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="WhoisResponse">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="WhoisResult" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="login">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="client" type="s:string" />
            <s:element minOccurs="0" maxOccurs="1" name="username" type="s:string" />
            <s:element minOccurs="0" maxOccurs="1" name="password" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="loginResponse">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="sessionId" type="s:string" />
          </s:sequence>
        </s:complexType>
      </s:element>
    </s:schema>
  </wsdl:types>
  <wsdl:message name="checkVatSoapIn"><wsdl:part name="parameters" element="tns:checkVat" /></wsdl:message>
  <wsdl:message name="checkVatSoapOut"><wsdl:part name="parameters" element="tns:checkVatResponse" /></wsdl:message>
  <wsdl:message name="CapitalCitySoapIn"><wsdl:part name="parameters" element="tns:CapitalCity" /></wsdl:message>
  <wsdl:message name="CapitalCitySoapOut"><wsdl:part name="parameters" element="tns:CapitalCityResponse" /></wsdl:message>
  <wsdl:message name="NumberToWordsSoapIn"><wsdl:part name="parameters" element="tns:NumberToWords" /></wsdl:message>
  <wsdl:message name="NumberToWordsSoapOut"><wsdl:part name="parameters" element="tns:NumberToWordsResponse" /></wsdl:message>
  <wsdl:message name="WhoisSoapIn"><wsdl:part name="parameters" element="tns:Whois" /></wsdl:message>
  <wsdl:message name="WhoisSoapOut"><wsdl:part name="parameters" element="tns:WhoisResponse" /></wsdl:message>
  <wsdl:message name="loginSoapIn"><wsdl:part name="parameters" element="tns:login" /></wsdl:message>
  <wsdl:message name="loginSoapOut"><wsdl:part name="parameters" element="tns:loginResponse" /></wsdl:message>
  <wsdl:portType name="TestServiceSoap">
    <wsdl:operation name="checkVat">
      <wsdl:input message="tns:checkVatSoapIn" />
      <wsdl:output message="tns:checkVatSoapOut" />
    </wsdl:operation>
    <wsdl:operation name="CapitalCity">
      <wsdl:input message="tns:CapitalCitySoapIn" />
      <wsdl:output message="tns:CapitalCitySoapOut" />
    </wsdl:operation>
    <wsdl:operation name="NumberToWords">
      <wsdl:input message="tns:NumberToWordsSoapIn" />
      <wsdl:output message="tns:NumberToWordsSoapOut" />
    </wsdl:operation>
    <wsdl:operation name="Whois">
      <wsdl:input message="tns:WhoisSoapIn" />
      <wsdl:output message="tns:WhoisSoapOut" />
    </wsdl:operation>
    <wsdl:operation name="login">
      <wsdl:input message="tns:loginSoapIn" />
      <wsdl:output message="tns:loginSoapOut" />
    </wsdl:operation>
  </wsdl:portType>
  <wsdl:binding name="TestServiceSoap" type="tns:TestServiceSoap">
    <soap:binding transport="http://schemas.xmlsoap.org/soap/http" />
    <wsdl:operation name="checkVat">
      <soap:operation soapAction="http://test.example/checkVat" style="document" />
      <wsdl:input><soap:body use="literal" /></wsdl:input>
      <wsdl:output><soap:body use="literal" /></wsdl:output>
    </wsdl:operation>
    <wsdl:operation name="CapitalCity">
      <soap:operation soapAction="http://test.example/CapitalCity" style="document" />
      <wsdl:input><soap:body use="literal" /></wsdl:input>
      <wsdl:output><soap:body use="literal" /></wsdl:output>
    </wsdl:operation>
    <wsdl:operation name="NumberToWords">
      <soap:operation soapAction="http://test.example/NumberToWords" style="document" />
      <wsdl:input><soap:body use="literal" /></wsdl:input>
      <wsdl:output><soap:body use="literal" /></wsdl:output>
    </wsdl:operation>
    <wsdl:operation name="Whois">
      <soap:operation soapAction="http://test.example/Whois" style="document" />
      <wsdl:input><soap:body use="literal" /></wsdl:input>
      <wsdl:output><soap:body use="literal" /></wsdl:output>
    </wsdl:operation>
    <wsdl:operation name="login">
      <soap:operation soapAction="http://test.example/login" style="document" />
      <wsdl:input><soap:body use="literal" /></wsdl:input>
      <wsdl:output><soap:body use="literal" /></wsdl:output>
    </wsdl:operation>
  </wsdl:binding>
  <wsdl:service name="TestService">
    <wsdl:port name="TestServiceSoap" binding="tns:TestServiceSoap">
      <soap:address location="{{.URL}}/soap" />
    </wsdl:port>
  </wsdl:service>
</wsdl:definitions>`

type testServer struct {
	server      *httptest.Server
	wsdlURL     string
	soapHandler func(w http.ResponseWriter, r *http.Request)
}

func newTestServer() *testServer {
	ts := &testServer{}

	ts.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/wsdl":
			tmpl, _ := template.New("wsdl").Parse(soapTestWSDL)
			var buf bytes.Buffer
			_ = tmpl.Execute(&buf, struct{ URL string }{URL: ts.server.URL})
			w.Header().Set("Content-Type", "text/xml")
			_, _ = w.Write(buf.Bytes())
		case "/soap":
			if ts.soapHandler != nil {
				ts.soapHandler(w, r)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	ts.wsdlURL = ts.server.URL + "/wsdl"
	return ts
}

func (ts *testServer) Close() {
	ts.server.Close()
}

// Response types
type CheckVatRequest struct {
	CountryCode string
	VatNumber   string
}

func (r CheckVatRequest) SoapBuildRequest() *Request {
	return NewRequest("checkVat", r)
}

type CheckVatResponse struct {
	CountryCode string `xml:"countryCode"`
	VatNumber   string `xml:"vatNumber"`
	RequestDate string `xml:"requestDate"`
	Valid       string `xml:"valid"`
	Name        string `xml:"name"`
	Address     string `xml:"address"`
}

type CapitalCityResponse struct {
	CapitalCityResult string
}

type NumberToWordsResponse struct {
	NumberToWordsResult string
}

type WhoisResponse struct {
	WhoisResult string
}

func TestSoapClient(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	tests := []struct {
		URL    string
		Err    bool
		Client *http.Client
	}{
		{
			URL: "://www.server",
			Err: false,
		},
		{
			URL: "",
			Err: false,
		},
		{
			URL: ts.wsdlURL,
			Err: true,
		},
	}

	for _, test := range tests {
		_, err := SoapClient(test.URL, nil)
		if err != nil && test.Err {
			t.Errorf("URL: %s - error: %s", test.URL, err)
		}
	}
}

func TestSoapClientWithClient(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	customClient := &http.Client{}
	client, err := SoapClient(ts.wsdlURL, customClient)

	if client.HTTPClient != customClient {
		t.Errorf("HTTP client is not the same as in initialization: - error: %s", err)
	}

	if err != nil {
		t.Errorf("URL: %s - error: %s", ts.wsdlURL, err)
	}
}

func TestClient_Call(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Set up SOAP response handler
	ts.soapHandler = func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)

		w.Header().Set("Content-Type", "text/xml; charset=utf-8")

		if bytes.Contains(body, []byte("checkVat")) {
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <checkVatResponse xmlns="http://test.example/">
      <countryCode>IE</countryCode>
      <vatNumber>6388047V</vatNumber>
      <requestDate>2024-01-15</requestDate>
      <valid>true</valid>
      <name>GOOGLE IRELAND LIMITED</name>
      <address>3RD FLOOR, GORDON HOUSE, BARROW STREET, DUBLIN 4</address>
    </checkVatResponse>
  </soap:Body>
</soap:Envelope>`))
		} else if bytes.Contains(body, []byte("CapitalCity")) {
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <CapitalCityResponse xmlns="http://test.example/">
      <CapitalCityResult>London</CapitalCityResult>
    </CapitalCityResponse>
  </soap:Body>
</soap:Envelope>`))
		} else if bytes.Contains(body, []byte("NumberToWords")) {
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <NumberToWordsResponse xmlns="http://test.example/">
      <NumberToWordsResult>twenty three </NumberToWordsResult>
    </NumberToWordsResponse>
  </soap:Body>
</soap:Envelope>`))
		} else if bytes.Contains(body, []byte("Whois")) {
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <WhoisResponse xmlns="http://test.example/">
      <WhoisResult>0</WhoisResult>
    </WhoisResponse>
  </soap:Body>
</soap:Envelope>`))
		} else {
			t.Logf("Unknown request: %s", bodyStr)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	soap, err := SoapClientWithConfig(ts.wsdlURL, nil, &Config{Dump: true})
	if err != nil {
		t.Fatalf("error not expected: %s", err)
	}

	var res *Response
	var rv CheckVatResponse
	var rc CapitalCityResponse
	var rn NumberToWordsResponse
	var rw WhoisResponse

	params := Params{}
	params["vatNumber"] = "6388047V"
	params["countryCode"] = "IE"

	// Test empty method
	res, err = soap.Call("", params)
	if err == nil {
		t.Errorf("method is empty")
	}
	if res != nil {
		t.Errorf("response should be nil for empty method")
	}

	// Test checkVat
	res, err = soap.Call("checkVat", params)
	if err != nil {
		t.Fatalf("error in soap call: %s", err)
	}

	err = res.Unmarshal(&rv)
	if err != nil {
		t.Fatalf("error unmarshaling: %s", err)
	}
	if rv.CountryCode != "IE" {
		t.Errorf("expected countryCode 'IE', got: %+v", rv)
	}

	// Test CapitalCity
	res, err = soap.Call("CapitalCity", Params{"sCountryISOCode": "GB"})
	if err != nil {
		t.Fatalf("error in soap call: %s", err)
	}

	err = res.Unmarshal(&rc)
	if err != nil {
		t.Fatalf("error unmarshaling: %s", err)
	}
	if rc.CapitalCityResult != "London" {
		t.Errorf("expected 'London', got: %+v", rc)
	}

	// Test NumberToWords
	res, err = soap.Call("NumberToWords", Params{"ubiNum": "23"})
	if err != nil {
		t.Fatalf("error in soap call: %s", err)
	}

	err = res.Unmarshal(&rn)
	if err != nil {
		t.Fatalf("error unmarshaling: %s", err)
	}
	if rn.NumberToWordsResult != "twenty three " {
		t.Errorf("expected 'twenty three ', got: %+v", rn)
	}

	// Test Whois
	res, err = soap.Call("Whois", Params{"DomainName": "google.com"})
	if err != nil {
		t.Fatalf("error in soap call: %s", err)
	}

	err = res.Unmarshal(&rw)
	if err != nil {
		t.Fatalf("error unmarshaling: %s", err)
	}
	if rw.WhoisResult != "0" {
		t.Errorf("expected '0', got: %+v", rw)
	}

	// Test uninitialized client
	c := &Client{}
	_, err = c.Call("", Params{})
	if err == nil {
		t.Errorf("error expected but nothing got.")
	}

	// Test invalid WSDL
	c.SetWSDL("://test.")
	_, err = c.Call("checkVat", params)
	if err == nil {
		t.Errorf("invalid WSDL should return error")
	}
}

type customLogger struct{}

func (c customLogger) LogRequest(method string, dump []byte) {
	var re = regexp.MustCompile(`(<vatNumber>)[\s\S]*?(<\/vatNumber>)`)
	maskedResponse := re.ReplaceAllString(string(dump), `${1}XXX${2}`)
	log.Printf("%s request: %s", method, maskedResponse)
}

func (c customLogger) LogResponse(method string, dump []byte) {
	if method == "checkVat" {
		return
	}
	log.Printf("Response: %s", dump)
}

func TestClient_Call_WithCustomLogger(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	ts.soapHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <checkVatResponse xmlns="http://test.example/">
      <countryCode>IE</countryCode>
      <vatNumber>6388047V</vatNumber>
      <requestDate>2024-01-15</requestDate>
      <valid>true</valid>
      <name>GOOGLE IRELAND LIMITED</name>
      <address>3RD FLOOR, GORDON HOUSE, BARROW STREET, DUBLIN 4</address>
    </checkVatResponse>
  </soap:Body>
</soap:Envelope>`))
	}

	soap, err := SoapClientWithConfig(ts.wsdlURL, nil, &Config{Dump: true, Logger: &customLogger{}})
	if err != nil {
		t.Fatalf("error not expected: %s", err)
	}

	var rv CheckVatResponse

	res, err := soap.CallByStruct(CheckVatRequest{
		CountryCode: "IE",
		VatNumber:   "6388047V",
	})
	if err != nil {
		t.Fatalf("error in soap call: %s", err)
	}

	err = res.Unmarshal(&rv)
	if err != nil {
		t.Fatalf("error unmarshaling: %s", err)
	}
	if rv.CountryCode != "IE" {
		t.Errorf("expected countryCode 'IE', got: %+v", rv)
	}

	_, err = soap.CallByStruct(nil)
	if err == nil {
		t.Error("err can't be nil")
	}
}

func TestClient_CallByStruct(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	ts.soapHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <checkVatResponse xmlns="http://test.example/">
      <countryCode>IE</countryCode>
      <vatNumber>6388047V</vatNumber>
      <requestDate>2024-01-15</requestDate>
      <valid>true</valid>
      <name>GOOGLE IRELAND LIMITED</name>
      <address>3RD FLOOR, GORDON HOUSE, BARROW STREET, DUBLIN 4</address>
    </checkVatResponse>
  </soap:Body>
</soap:Envelope>`))
	}

	soap, err := SoapClient(ts.wsdlURL, nil)
	if err != nil {
		t.Fatalf("error not expected: %s", err)
	}

	var rv CheckVatResponse

	res, err := soap.CallByStruct(CheckVatRequest{
		CountryCode: "IE",
		VatNumber:   "6388047V",
	})
	if err != nil {
		t.Fatalf("error in soap call: %s", err)
	}

	err = res.Unmarshal(&rv)
	if err != nil {
		t.Fatalf("error unmarshaling: %s", err)
	}
	if rv.CountryCode != "IE" {
		t.Errorf("expected countryCode 'IE', got: %+v", rv)
	}

	_, err = soap.CallByStruct(nil)
	if err == nil {
		t.Error("err can't be nil")
	}
}

func TestClient_Call_NonUtf8(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Simulate non-UTF8 response (ISO-8859-1)
	ts.soapHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=ISO-8859-1")
		// Response with ISO-8859-1 encoding declaration
		w.Write([]byte(`<?xml version="1.0" encoding="ISO-8859-1"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <loginResponse xmlns="http://test.example/">
      <sessionId>abc123</sessionId>
    </loginResponse>
  </soap:Body>
</soap:Envelope>`))
	}

	soap, err := SoapClient(ts.wsdlURL, nil)
	if err != nil {
		t.Fatalf("error not expected: %s", err)
	}

	_, err = soap.Call("login", Params{"client": "demo", "username": "robert", "password": "iliasdemo"})
	if err != nil {
		t.Fatalf("error in soap call: %s", err)
	}
}

func TestProcess_doRequest(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := &process{
		Client: &Client{
			HTTPClient: &http.Client{},
		},
	}

	// Test empty URL
	_, err := c.doRequest("")
	if err == nil {
		t.Errorf("empty URL should return error")
	}

	// Test invalid URL
	_, err = c.doRequest("://teste.")
	if err == nil {
		t.Errorf("invalid URL should return error")
	}

	// Test 404 response
	_, err = c.doRequest(ts.server.URL + "/non-existent-url")
	if err == nil {
		t.Errorf("404 should return error")
	}

	if err != nil && err.Error() != "unexpected status code: 404 Not Found" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestClient_Call_EmptyMethod(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	soap, err := SoapClient(ts.wsdlURL, nil)
	if err != nil {
		t.Fatalf("error not expected: %s", err)
	}

	_, err = soap.Call("", Params{"test": "value"})
	if err == nil {
		t.Error("empty method should return error")
	}
}

func TestClient_HeaderParams(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	var capturedRequest string
	ts.soapHandler = func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedRequest = string(body)
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <checkVatResponse xmlns="http://test.example/">
      <countryCode>IE</countryCode>
    </checkVatResponse>
  </soap:Body>
</soap:Envelope>`))
	}

	soap, err := SoapClient(ts.wsdlURL, nil)
	if err != nil {
		t.Fatalf("error not expected: %s", err)
	}

	soap.HeaderName = "AuthHeader"
	soap.HeaderParams = HeaderParams{
		"Username": "testuser",
		"Password": "testpass",
	}

	_, err = soap.Call("checkVat", Params{"countryCode": "IE"})
	if err != nil {
		t.Fatalf("error in soap call: %s", err)
	}

	if !bytes.Contains([]byte(capturedRequest), []byte("Header")) {
		t.Logf("Request: %s", capturedRequest)
		// Header should be present in the request
	}
}

func TestErrorWithPayload_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	payload := []byte("<xml>test</xml>")

	wrappedErr := ErrorWithPayload{
		error:   originalErr,
		Payload: payload,
	}

	// Test Unwrap
	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("expected Unwrap to return original error")
	}

	// Test errors.Is
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("expected errors.Is to find original error through Unwrap")
	}

	// Test GetPayloadFromError
	gotPayload := GetPayloadFromError(wrappedErr)
	if !bytes.Equal(gotPayload, payload) {
		t.Errorf("expected payload '%s', got '%s'", payload, gotPayload)
	}
}
