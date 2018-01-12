package api

type Clusters struct {
	Cluster []Cluster `xml:"cluster"`
}

type Cluster struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name"`
}
