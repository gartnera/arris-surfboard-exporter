package lib

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Scraper struct {
	serverBase string
	jar        *cookiejar.Jar
	hClient    *http.Client

	creds string
	token string
}

func NewScraper(serverBase, creds string) (*Scraper, error) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return &Scraper{
		serverBase: serverBase,
		jar:        jar,
		hClient:    client,
		creds:      creds,
	}, err
}

func (c *Scraper) ensureLogin() error {
	loginPath := fmt.Sprintf("/cmconnectionstatus.html?login_%s", c.creds)
	loginUrl := c.serverBase + loginPath
	req, err := http.NewRequest("GET", loginUrl, nil)
	if err != nil {
		return err
	}
	req.Header = http.Header{
		"Authorization": []string{fmt.Sprintf("Basic %s", c.creds)},
		"Content-Type":  []string{"application/x-www-form-urlencoded; charset=utf-8"},
	}
	resp, err := c.hClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got nonzero status code when logging in: %d", resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: %w", err)
	}
	token := string(bodyBytes)
	if len(token) < 10 || len(token) > 100 {
		return fmt.Errorf("got unexpected token: %s", token)
	}
	c.token = token
	return nil
}

func (c *Scraper) GetConnectionStatus() (*ConnectionStatusResult, error) {
	err := c.ensureLogin()
	if err != nil {
		return nil, err
	}
	loginPath := fmt.Sprintf("/cmconnectionstatus.html?ct_%s", c.token)
	loginUrl := c.serverBase + loginPath
	req, err := http.NewRequest("GET", loginUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{
		"Authorization": []string{fmt.Sprintf("Basic %s", c.creds)},
	}
	resp, err := c.hClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got nonzero status code when logging in: %d", resp.StatusCode)
	}

	parser := ConnectionStatusParser{}
	err = parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	return &parser.results, nil
}

func (c *Scraper) runInner(upstreamCb func(*UpstreamBondedChannel), downstreamCb func(*DownstreamBondedChannel)) error {
	result, err := c.GetConnectionStatus()
	if err != nil {
		return err
	}
	for _, res := range result.UpstreamBondedChannel {
		upstreamCb(res)
	}
	for _, res := range result.DownstreamBondedChannel {
		downstreamCb(res)
	}
	return nil
}

func (c *Scraper) Run(upstreamCb func(*UpstreamBondedChannel), downstreamCb func(*DownstreamBondedChannel)) error {
	for {
		err := c.runInner(upstreamCb, downstreamCb)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(10 * time.Second)
	}
}
