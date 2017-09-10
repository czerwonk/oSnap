package api

import "encoding/xml"

type Snapshot struct {
	XMLName            xml.Name `xml:"snapshot"`
	Description        string   `xml:"description"`
	PersistMemoryState bool     `xml:"persist_memorystate"`
}
