package api

type Vms struct {
	Vm []Vm `xml:"vm"`
}

type Vm struct {
	Id      string `xml:"id,attr"`
	Name    string `xml:"name"`
	Cluster struct {
		Id string `xml:"id,attr"`
	} `xml:"cluster,omitempty"`
}
