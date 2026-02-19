package soap

import "encoding/xml"

// Envelope represents the root <soapenv:Envelope> tag.
type Envelope struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	Xmlns   string   `xml:"xmlns:soapenv,attr"`
	Web     string   `xml:"xmlns:web,attr"`
	Header  *Header  `xml:"soapenv:Header,omitempty"`
	Body    Body
}

// Header is empty but required by the protocol.
type Header struct {
	XMLName xml.Name `xml:"soapenv:Header"`
}

// Body contains the actual function call.
type Body struct {
	XMLName xml.Name `xml:"soapenv:Body"`
	Process Process  `xml:"web:process"`
}

// Process maps to the Movitel USSD Push API format.
type Process struct {
	Username string `xml:"username"`
	Password string `xml:"password"`
	Wscode   string `xml:"wscode"`
	Param    string `xml:"param"`
	RawData  string `xml:"rawData,omitempty"`
}

// ResponseEnvelope represents the envelope returned by Movitel.
type ResponseEnvelope struct {
	XMLName xml.Name     `xml:"Envelope"`
	Body    ResponseBody `xml:"Body"`
}

type ResponseBody struct {
	ProcessResp ProcessResponse `xml:"processResponse"`
}

// ProcessResponse contains the immediate result with escaped XML.
type ProcessResponse struct {
	Return   string `xml:"return"`
	Original string `xml:"original"`
}

// DetailResponse represents the parsed content of the <return> tag.
type DetailResponse struct {
	ErrorCode       string `xml:"errorCode"`
	Message         string `xml:"message"`
	TransID         string `xml:"transId"`
	Balance         string `xml:"balance"`
	OrgResponseCode string `xml:"orgResponseCode"`
}

// NewEnvelope creates a properly formatted SOAP Envelope.
func NewEnvelope(username, password, wscode, paramXML string) *Envelope {
	return &Envelope{
		Xmlns: "http://schemas.xmlsoap.org/soap/envelope/",
		Web:   "http://webservice.com/",
		Body: Body{
			Process: Process{
				Username: username,
				Password: password,
				Wscode:   wscode,
				Param:    paramXML,
			},
		},
	}
}
