package exporter

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
	log "github.com/sirupsen/logrus"
)

type error interface {
	Error() string
}

type Fetcher struct {
	saltUrl      string
	saltUser     string
	saltPassword string
	saltToken    string
}

type Minions struct {
	list  []string
	count int
}

type Masters struct {
	list   []string
	count  int
	status map[string]bool
}

/*
 * Builder of new Fetcher pseudo-object
 * Get the saltstack url, user, password
 * Return the fetch structure initialized
 */
func NewFetcher(saltUrl string, saltUser string, saltPassword string) *Fetcher {
	return &Fetcher{
		saltUrl:      saltUrl,
		saltUser:     saltUser,
		saltPassword: saltPassword,
		saltToken:    "",
	}
}

/*
 * Function associated to the Fetcher pseudo-object to login into the saltstack API
 * Return an error if it's occuring, nil otherwise
 */
func (f *Fetcher) Login() error {
	client := http.Client{
		Timeout: time.Second * 2,
	}

	form := url.Values{}
	form.Add("username", f.saltUser)
	form.Add("password", f.saltPassword)
	form.Add("eauth", "pam")

	req, _ := http.NewRequest(http.MethodPost, f.saltUrl+"/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, getErr := client.Do(req)
	if getErr != nil {
		return fmt.Errorf("error during request: %v", getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return fmt.Errorf("error reading result's body: %v", readErr)
	}

	if res.Status[:3] != "200" {
		return fmt.Errorf("unknown error: %v", res.Status)
	}

	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	f.saltToken = strings.Trim(jsonParsed.Path("return.0.token").Data().(string), "[]\"")
	log.WithFields(log.Fields{
		"token": f.saltToken,
	}).Debug("Displaying API token")

	return nil
}

/*
 * Helper to retreive JSON from url and returning the GABS object
 * Take an `url` in parameter
 * Return a GABS container (json parsed) and/or error if it happens
 */
func (f *Fetcher) getJson(url string) (*gabs.Container, error) {
	client := http.Client{}

	if f.saltToken == "" {
		err := f.Login()
		if err != nil {
			return gabs.New(), fmt.Errorf("error during login: %v", err)
		}
	}

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("X-Auth-Token", ""+f.saltToken)

	res, getErr := client.Do(req)
	if getErr != nil {
		return gabs.New(), fmt.Errorf("error during request: %v", getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return gabs.New(), fmt.Errorf("error reading result's body: %v", readErr)
	}

	if res.Status[:3] == "401" {
		return gabs.New(), fmt.Errorf("unauthorized, token must be refreshed: %v", res.Status)
	}

	if res.Status[:3] != "200" {
		return gabs.New(), fmt.Errorf("unknown error: %v", res.Status)
	}

	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		return gabs.New(), fmt.Errorf("error parsing JSON: %v", err)
	}

	return jsonParsed, nil
}

/*
 * Function associated to the Fetcher pseudo-object that fetch information about minions
 * Return a couple of Minion struct and error if it happens, nil otherwise
 */
func (f *Fetcher) Minions() (*Minions, error) {
	var minions = Minions{}

	jsonParsed, err := f.getJson(f.saltUrl + "/minions")

	if err != nil {
		return &minions, fmt.Errorf("error parsing JSON: %v", err)
	}

	for _, ele := range jsonParsed.S("return").Children() {
		for key := range ele.ChildrenMap() {
			minions.list = append(minions.list, key)
		}
	}

	minions.count = len(minions.list)

	log.WithFields(log.Fields{
		"count": minions.count,
		"list":  minions.list,
	}).Debug("displaying minions informations")
	return &minions, nil
}

/*
 * Function that return a list of minion masters
 * Get a list of minions in parameters
 * Return a list of master
 */
func (f *Fetcher) Masters() (*Masters, error) {
	var masters = Masters{}
	var umasters = make(map[string]bool)

	masters.status = make(map[string]bool)

	jsonParsed, err := f.getJson(f.saltUrl + "/minions")

	if err != nil {
		return &masters, fmt.Errorf("error parsing JSON: %v", err)
	}

	for _, ele := range jsonParsed.S("return").Children() {
		for _, elem := range ele.ChildrenMap() {
			master := elem.Path("master").Data().(string)
			if !umasters[master] {
				umasters[master] = true
				masters.list = append(masters.list, master)
			}
		}
	}

	masters.count = len(masters.list)

	for _, master := range masters.list {
		// TODO: check the master status
		masters.status[master] = true
	}

	log.WithFields(log.Fields{
		"count":  masters.count,
		"list":   masters.list,
		"status": masters.status,
	}).Debug("displaying master informations")
	return &masters, nil
}

// Faire une fonction qui retourne le status d'un master (parametres)
// Faire une fonction qui retourne l'etat du dernier highstate/apply pour un agent
// Faire une function qui retourne le compte de highstate/apply total par agent
// c.f. les possibiltés avec l'API
