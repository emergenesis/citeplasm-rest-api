package main

import (
    "log"
    "godis"
    "strings"
    "strconv"
)

// Provider represents a provider of Resources.
type Provider struct {

    // Identifier is the unique ID of the resource.
    Identifier string `json:"-"`

    // Name is a friendly representation of the resource.
    Name string `json:"name"`

    // Icon is an URL to a 24x24 icon representative of the Provider.
    Icon string `json:"icon"`

    // Logo is an URL to a larger image representative of the Provider.
    Logo string `json:"logo"`

    // Description is a long-form explanation of the Provider.
    Description string `json:"descr"`

    // db is an internal pointer to the database connection.
    db *godis.Client `json:"-"`
}

// NewProvider creates a new Provider.
func NewProvider( db *godis.Client, name string ) *Provider {
    var p Provider

    // get the next available provider ID
    i64, err := db.Incr("nxProvId")
    if err != nil {
        log.Fatal("Could not INCR nxProvId!")
    }

    // build the Provider
    p.Identifier = strconv.FormatInt(i64, 10)
    p.Name = name
    p.db = db

    // return a pointer to the Provider
    return &p
}

// GetKey returns the database key for this Provider.
func (p *Provider) GetKey() string {
    return "prov:" + p.Identifier
}

// Db returns a pointer to the database client.
func (p *Provider) Db() *godis.Client {
    return p.db
}

// Id returns the ID of this Provider.
func (p *Provider) Id() string {
    return p.Identifier
}

// Label returns the Name of this Provider.
func (p *Provider) Label() string {
    return p.Name
}

// Uri returns the URI of this Provider within this API.
func (p *Provider) Uri() string {
    return "/providers/" + p.Identifier
}

// GetProviders returns an array of the first 10 Providers.
// FIXME support different result counts
func GetProviders (db *godis.Client) []Resource {
    var providers []Resource

    // fetch the Providers
    r, err := db.Lrange("idx:Provider", 0, 10)
    if err != nil {
        log.Fatal("Could not get providers.")
    }

    // conver the Reply into a []string of "Id|Label" representations
    s := r.StringArray()

    // loop through all results, creating resources
    for i := 0; i < len(s); i++ {
        // split the "Id|Label"
        vals := strings.SplitN(s[i], "|", 2)
        if len(vals) != 2 {
            log.Fatalf("For some reason, the index was improperly formatted: %s", s[i])
        }

        // build the Provider
        var p Provider
        p.Identifier = vals[0]
        p.Name = vals[1]

        // create a resource from the provider and add it to the array to return
        providers = append(providers, NewResource(&p))
    }

    // return the array of providers
    return providers
}
