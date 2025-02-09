package metainfo

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"tgragnato.it/magnetico/types/infohash"
)

// Magnet link components.
type Magnet struct {
	InfoHash    infohash.T // Expected in this implementation
	Trackers    []string   // "tr" values
	DisplayName string     // "dn" value, if not empty
	Params      url.Values // All other values, such as "x.pe", "as", "xs" etc.
}

const btihPrefix = "urn:btih:"

func (m Magnet) String() string {
	// Deep-copy m.Params
	vs := make(url.Values, len(m.Params)+len(m.Trackers)+2)
	for k, v := range m.Params {
		vs[k] = append([]string(nil), v...)
	}

	for _, tr := range m.Trackers {
		vs.Add("tr", tr)
	}
	if m.DisplayName != "" {
		vs.Add("dn", m.DisplayName)
	}

	// Transmission and Deluge both expect "urn:btih:" to be unescaped. Deluge wants it to be at the
	// start of the magnet link. The InfoHash field is expected to be BitTorrent in this
	// implementation.
	u := url.URL{
		Scheme:   "magnet",
		RawQuery: "xt=" + btihPrefix + m.InfoHash.HexString(),
	}
	if len(vs) != 0 {
		u.RawQuery += "&" + vs.Encode()
	}
	return u.String()
}

// Deprecated: Use ParseMagnetUri.
var ParseMagnetURI = ParseMagnetUri

// ParseMagnetUri parses Magnet-formatted URIs into a Magnet instance
func ParseMagnetUri(uri string) (m Magnet, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("error parsing uri: %w", err)
		return
	}
	if u.Scheme != "magnet" {
		err = fmt.Errorf("unexpected scheme %q", u.Scheme)
		return
	}
	q := u.Query()
	gotInfohash := false
	for _, xt := range q["xt"] {
		if gotInfohash {
			lazyAddParam(&m.Params, "xt", xt)
			continue
		}
		encoded, found := strings.CutPrefix(xt, btihPrefix)
		if !found {
			lazyAddParam(&m.Params, "xt", xt)
			continue
		}
		m.InfoHash, err = parseEncodedV1Infohash(encoded)
		if err != nil {
			err = fmt.Errorf("error parsing v1 infohash %q: %w", xt, err)
			return
		}
		gotInfohash = true
	}
	if !gotInfohash {
		err = errors.New("missing v1 infohash")
		return
	}
	q.Del("xt")
	m.DisplayName = popFirstValue(q, "dn")
	m.Trackers = q["tr"]
	q.Del("tr")
	copyParams(&m.Params, q)
	return
}

func parseEncodedV1Infohash(encoded string) (ih infohash.T, err error) {
	decode := func() func(dst, src []byte) (int, error) {
		switch len(encoded) {
		case 40:
			return hex.Decode
		case 32:
			return base32.StdEncoding.Decode
		}
		return nil
	}()
	if decode == nil {
		err = fmt.Errorf("unhandled xt parameter encoding (encoded length %d)", len(encoded))
		return
	}
	n, err := decode(ih[:], []byte(encoded))
	if err != nil {
		err = fmt.Errorf("error decoding xt: %w", err)
		return
	}
	if n != 20 {
		return infohash.T{}, errors.New("decoded xt length != 20")
	}
	return
}

func popFirstValue(vs url.Values, key string) string {
	sl := vs[key]
	switch len(sl) {
	case 0:
		return ""
	case 1:
		vs.Del(key)
		return sl[0]
	default:
		vs[key] = sl[1:]
		return sl[0]
	}
}
