package geoip

import (
	"fmt"
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

type Locator struct {
	reader *geoip2.Reader
	log    *logrus.Logger
}

func Open(path string, log *logrus.Logger) (*Locator, error) {
	if path == "" {
		return nil, nil
	}

	reader, err := geoip2.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open GeoIP database: %w", err)
	}

	return &Locator{reader: reader, log: log}, nil
}

func (l *Locator) CountryCode(ip string) string {
	if l == nil || l.reader == nil {
		return ""
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}

	record, err := l.reader.Country(parsed)
	if err != nil {
		if l.log != nil {
			l.log.Warnf("GeoIP lookup failed: %v", err)
		}
		return ""
	}
	return record.Country.IsoCode
}

func (l *Locator) Close() error {
	if l == nil || l.reader == nil {
		return nil
	}
	return l.reader.Close()
}
