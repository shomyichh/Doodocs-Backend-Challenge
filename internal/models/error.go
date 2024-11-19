package models

import "encoding/xml"

type ErrorResponse struct {
	XMLName     xml.Name `xml:"error"`
	Code        int      `xml:"code"`
	Message     string   `xml:"message"`
	Description string   `xml:"description"`
}
