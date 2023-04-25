package lib

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type UpstreamBondedChannel struct {
	Channel       string
	ChannelID     string
	LockStatus    string
	USChannelType string
	Frequency     int
	Width         int
	Power         float64
}

func (c *UpstreamBondedChannel) Labels() map[string]string {
	return map[string]string{
		"channel":         c.Channel,
		"channel_id":      c.ChannelID,
		"us_channel_type": c.USChannelType,
	}
}

type DownstreamBondedChannel struct {
	ChannelID      string
	LockStatus     string
	Modulation     string
	Frequency      int
	Power          float64
	SNR            float64
	Corrected      int
	Uncorrectables int
}

func (c *DownstreamBondedChannel) Labels() map[string]string {
	return map[string]string{
		"channel_id": c.ChannelID,
		"modulation": c.Modulation,
	}
}

func stringToIntStrip(val string) int {
	res, _ := strconv.Atoi(strings.Split(val, " ")[0])
	return res
}

func stringToFloatStrip(val string) float64 {
	res, _ := strconv.ParseFloat(strings.Split(val, " ")[0], 64)
	return res
}

type ConnectionStatusResult struct {
	UpstreamBondedChannel   []*UpstreamBondedChannel
	DownstreamBondedChannel []*DownstreamBondedChannel
}

type ConnectionStatusParser struct {
	results ConnectionStatusResult
}

func (p *ConnectionStatusParser) ParseBytes(buf []byte) error {
	return p.Parse(bytes.NewReader(buf))
}

func (p *ConnectionStatusParser) handleUpstreamTable(s *goquery.Selection) {
	s.Children().Find("tr").Each(func(i int, s *goquery.Selection) {
		// skip header
		if i < 2 {
			return
		}
		res := &UpstreamBondedChannel{}
		s.Children().Each(func(j int, s *goquery.Selection) {
			val := s.Text()
			switch j {
			case 0:
				res.Channel = val
			case 1:
				res.ChannelID = val
			case 2:
				res.LockStatus = val
			case 3:
				res.USChannelType = val
			case 4:
				res.Frequency = stringToIntStrip(val)
			case 5:
				res.Width = stringToIntStrip(val)
			case 6:
				res.Power = stringToFloatStrip(val)
			}
		})
		p.results.UpstreamBondedChannel = append(p.results.UpstreamBondedChannel, res)
	})
}

func (p *ConnectionStatusParser) handleDownstreamTable(s *goquery.Selection) {
	s.Children().Find("tr").Each(func(i int, s *goquery.Selection) {
		// skip header
		if i < 2 {
			return
		}
		res := &DownstreamBondedChannel{}
		s.Children().Each(func(j int, s *goquery.Selection) {
			val := s.Text()
			switch j {
			case 0:
				res.ChannelID = val
			case 1:
				res.LockStatus = val
			case 2:
				res.Modulation = val
			case 3:
				res.Frequency = stringToIntStrip(val)
			case 4:
				res.Power = stringToFloatStrip(val)
			case 5:
				res.SNR = stringToFloatStrip(val)
			case 6:
				res.Corrected = stringToIntStrip(val)
			case 7:
				res.Uncorrectables = stringToIntStrip(val)
			}
		})
		p.results.DownstreamBondedChannel = append(p.results.DownstreamBondedChannel, res)
	})
}

func (p *ConnectionStatusParser) Parse(reader io.Reader) error {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return fmt.Errorf("unable to parse document: %w", err)
	}
	doc.Find(".simpleTable").Each(func(i int, s *goquery.Selection) {
		title := s.Find("strong").First().Text()
		if strings.Contains(title, "Upstream Bonded Channels") {
			p.handleUpstreamTable(s)
		}
		if strings.Contains(title, "Downstream Bonded Channels") {
			p.handleDownstreamTable(s)
		}
	})
	return nil
}
