package rtorrent

import (
	"fmt"
	"github.com/ConorNevin/xmlrpc"
	"net/http"
)

type RTorrent struct {
	xmlrpcClient *xmlrpc.Client
}

type Torrent struct {
	Hash      string
	Name      string
	Path      string
	Size      int64
	Label     string
	Completed bool
	Ratio     float64
}

type View string

const (
	// ViewMain represents the "main" view, containing all torrents
	ViewMain View = "main"
	// ViewStarted represents the "started" view, containing only torrents that have been started
	ViewStarted View = "started"
	// ViewStopped represents the "stopped" view, containing only torrents that have been stopped
	ViewStopped View = "stopped"
	// ViewHashing represents the "hashing" view, containing only torrents that are currently hashing
	ViewHashing View = "hashing"
	// ViewSeeding represents the "seeding" view, containing only torrents that are currently seeding
	ViewSeeding View = "seeding"
)

func New(addr string) *RTorrent {
	client, _ := xmlrpc.NewClient(addr, nil)

	return &RTorrent{
		xmlrpcClient: client,
	}
}

func NewWithCredentials(addr string, credentials *Credentials) *RTorrent {
	basicAuthRT := NewBasicAuthRoundTripper(http.DefaultTransport, credentials)

	client, _ := xmlrpc.NewClient(addr, basicAuthRT)

	return &RTorrent{
		xmlrpcClient: client,
	}
}

func (r *RTorrent) Name() (string, error) {
	var result string

	err := r.xmlrpcClient.Call("get_name", nil, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (r *RTorrent) IP() (string, error) {
	var result string

	err := r.xmlrpcClient.Call("get_ip", nil, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (r *RTorrent) GetTorrents(view View) ([]Torrent, error) {
	args := []interface{}{string(view), "d.get_name=", "d.get_size_bytes=", "d.get_hash=", "d.get_custom1=", "d.get_base_path=", "d.is_active=", "d.get_complete=", "d.get_ratio="}

	var results []interface{}
	err := r.xmlrpcClient.Call("d.multicall", args, &results)

	var torrents []Torrent
	if err != nil {
		return torrents, fmt.Errorf("Failed to fetch torrents: %v", err)
	}

	for _, result := range results {
		torrentData := result.([]interface{})
		torrents = append(torrents, Torrent{
			Hash:      toString(torrentData[2]),
			Name:      toString(torrentData[0]),
			Path:      toString(torrentData[4]),
			Size:      torrentData[1].(int64),
			Label:     toString(torrentData[3]),
			Completed: torrentData[6].(int64) > 0,
			Ratio:     float64(torrentData[7].(int64)) / float64(1000),
		})
	}

	return torrents, nil
}

func toString(val interface{}) string {
	switch val.(type) {
	case string:
		return val.(string)
	case nil:
		return ""
	default:
		panic("Panic!")
	}
}

type Credentials struct {
	Username string
	Password string
}

type BasicAuthRoundTripper struct {
	RoundTripper http.RoundTripper
	Credentials  *Credentials
}

func NewBasicAuthRoundTripper(roundTripper http.RoundTripper, creds *Credentials) *BasicAuthRoundTripper {
	return &BasicAuthRoundTripper{
		RoundTripper: roundTripper,
		Credentials:  creds,
	}
}

func (r *BasicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(r.Credentials.Username, r.Credentials.Password)

	return r.RoundTripper.RoundTrip(req)
}
