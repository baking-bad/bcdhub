package sitemap

import (
	"encoding/xml"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
)

// Sitemap is awesome
type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
	modDate string
}

// URL is awesome
type URL struct {
	XMLName  xml.Name `xml:"url"`
	Location string   `xml:"loc"`
	LastMod  string   `xml:"lastmod"`
}

// New - creates new sitemap
func New() *Sitemap {
	return &Sitemap{
		Xmlns:   "http://www.sitemaps.org/schemas/sitemap/0.9",
		modDate: time.Now().Format("2006-01-02"),
	}
}

// AddLocation - adds new location to sitemap
func (s *Sitemap) AddLocation(location string) {
	s.URLs = append(s.URLs, URL{Location: location, LastMod: s.modDate})
}

// SaveToFile - save sitemap to path
func (s *Sitemap) SaveToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	xmlWriter := io.Writer(file)
	if _, err = xmlWriter.Write([]byte(xml.Header)); err != nil {
		return err
	}

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("", "  ")
	if err := enc.Encode(s); err != nil {
		return errors.Errorf("encode error: %v", err)
	}

	return nil
}
