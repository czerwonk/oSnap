package api

type Vms struct {
	Vm []Vm `xml:"vm"`
}

type Vm struct {
	Id      string `xml:"id,attr,omitempty"`
	Name    string `xml:"name,omitempty"`
	Cluster struct {
		Id string `xml:"id,attr"`
	} `xml:"cluster,omitempty"`
}
