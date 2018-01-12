package api

type Vms struct {
	VM []Vm `xml:"vm"`
}

type Vm struct {
	ID      string `xml:"id,attr,omitempty"`
	Name    string `xml:"name,omitempty"`
	Cluster struct {
		ID string `xml:"id,attr"`
	} `xml:"cluster,omitempty"`
}
