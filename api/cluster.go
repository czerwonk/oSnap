package api

type Clusters struct {
	Cluster []Cluster `xml:"cluster"`
}

type Cluster struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"name"`
}
