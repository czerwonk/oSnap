package api

import "encoding/xml"

type Snapshots struct {
	Snapshot []Snapshot `xml:"snapshot"`
}

type Snapshot struct {
	XMLName            xml.Name `xml:"snapshot"`
	Id                 string   `xml:"id,attr,omitempty"`
	Description        string   `xml:"description"`
	PersistMemoryState bool     `xml:"persist_memorystate"`
	Status             string   `xml:"snapshot_status,omitempty"`
	Vm                 Vm       `xml:"vm"`
}
