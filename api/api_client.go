package api

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type ApiClient struct {
	url          string
	user         string
	pass         string
	insecureCert bool
	client       *http.Client
}

func NewClient(url, user, pass string, insecureCert bool) *ApiClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureCert},
	}
	c := &http.Client{Transport: tr}

	return &ApiClient{url: url, user: user, pass: pass, insecureCert: insecureCert, client: c}
}

func (c *ApiClient) GetVms(clusterId string) ([]Vm, error) {
	clusterId, err := c.getClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	vms := Vms{}
	err = c.sendAndParse("vms", "GET", &vms)
	if err != nil {
		return nil, err
	}

	res := make([]Vm, 0)
	for _, v := range vms.Vm {
		if v.Cluster.Id == clusterId {
			res = append(res, v)
		}
	}

	return res, nil
}

func (c *ApiClient) getClusterId(name string) (string, error) {
	clusters := Clusters{}
	err := c.sendAndParse(fmt.Sprintf("clusters?search=%s", name), "GET", &clusters)
	if err != nil {
		return "", err
	}

	for _, cluster := range clusters.Cluster {
		if cluster.Name == name {
			return cluster.Id, nil
		}
	}

	return "", errors.New("Unknown cluster " + name)
}

func (c *ApiClient) CreateSnapshot(vmId string) error {
	s := &Snapshot{Description: "Simple oVirt Snapshot", PersistMemoryState: false}
	b, err := xml.Marshal(s)
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)
	_, err = c.sendRequest(fmt.Sprintf("vms/%s/snapshots", vmId), "POST", r)
	return err
}

func (c *ApiClient) sendAndParse(path, method string, res interface{}) error {
	b, err := c.sendRequest(path, method, nil)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(b, res)
	return err
}

func (c *ApiClient) sendRequest(path, method string, body io.Reader) ([]byte, error) {
	uri := strings.Trim(c.url, "/") + "/" + strings.Trim(path, "/")
	log.Println(method, uri)

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/xml")
	req.SetBasicAuth(c.user, c.pass)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, errors.New(fmt.Sprintf(resp.Status))
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
