package main

import (
    "log"
    "godis"
    "strings"
    "strconv"
)

// Provider represents a provider of Resources
type Provider struct {
    Identifier string `json:"-"`
    Name string `json:"name"`
    Icon string `json:"icon"`
    Logo string `json:"logo"`
    Description string `json:"descr"`
    db *godis.Client `json:"-"`
}

// NewProvider creates a new Provider
func NewProvider( db *godis.Client, name string ) *Provider {
    var p Provider

    i64, err := db.Incr("nxProvId")
    if err != nil {
        log.Fatal("Could not INCR nxProvId!")
    }

    p.Identifier = strconv.FormatInt(i64, 10)
    p.Name = name
    p.db = db

    return &p
}

// GetKey returns the database key for this Provider
func (p *Provider) GetKey() string {
    return "prov:" + p.Identifier
}

// Db returns a pointer to the database client
func (p *Provider) Db() *godis.Client {
    return p.db
}

// Id returns the ID of this Provider
func (p *Provider) Id() string {
    return p.Identifier
}

// Label returns the Name of this Provider
func (p *Provider) Label() string {
    return p.Name
}

// Uri returns the URI of this Provider
func (p *Provider) Uri() string {
    return "/providers/" + p.Identifier
}

// GetProvider
/*func GetProvider () Provider {
}*/

// GetProviders
func GetProviders (db *godis.Client) []Resource {
    // FIXME support different result counts

    var providers []Resource
    r, err := db.Lrange("idx:Provider", 0, 10)
    if err != nil {
        log.Fatal("Could not get providers.")
    }
    s := r.StringArray()

    // loop through all results, creating resources
    for i := 0; i < len(s); i++ {
        vals := strings.SplitN(s[i], "|", 2)
        if len(vals) != 2 {
            log.Fatalf("For some reason, the index was improperly formatted: %s", s[i])
        }

        var p Provider
        p.Identifier = vals[0]
        p.Name = vals[1]
        providers = append(providers, NewResource(&p))
    }

    return providers
}
