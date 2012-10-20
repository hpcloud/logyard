package stackato

import (
	"github.com/ActiveState/doozer"
	"io/ioutil"
	"strings"
)

// GetDoozerURI returns the doozer URI for the current Stackato
// cluster.
func GetDoozerURI() (string, error) {
	data, err := ioutil.ReadFile("/s/etc/doozer/doozer_uri")
	if err != nil {
		return "", err
	}
	uri := strings.SplitN(string(data), "=", 2)[1]
	return uri, nil
}

// NewDoozerClient returns the doozer connection to Stackato's doozer
// servers.
func NewDoozerClient(clientName string) (*doozer.Conn, int64, error) {
	uri, err := GetDoozerURI()
	if err != nil {
		return nil, 0, err
	}
	conn, err := doozer.Dial(uri)
	if err != nil {
		return nil, 0, err
	}

	headRev, err := conn.Rev()
	if err != nil {
		return nil, 0, err
	}

	// Set ephemeral file value
	_, err = conn.Set("/eph", headRev, []byte(clientName))
	if err != nil {
		return nil, 0, err
	}

	return conn, headRev, nil
}
