// Package neohttp implements a driver for neo4j via the http(s) interface.
// Use with package cypher by importing this package as _ and connecting using cypher.Connect("neohttp", ...)
package neohttp

import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tjbrockmeyer/cypher"
	"strings"
)

var supportedMajorVersions = []string{
	"4",
}

func init() {
	cypher.Register("neohttp", driver{})
}

type driver struct{}

func (d driver) Connect(uri, dbName, username, password string) (cypher.DB, error) {
	db := &database{
		uri:       uri,
		discovery: struct {
			BoltDirect string `json:"bolt_direct"`
			Cluster    string `json:"cluster"`
			TX         string `json:"transaction"`
			Version    string `json:"neo4j_version"`
			Edition    string `json:"neo4j_edition"`
		}{},
	}
	if username != "" {
		db.basicAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(username, ":", password)))
	}
	if err := db.connectWithRetry(4); err != nil {
		return nil, err
	}
	if !versionIsSupported(strings.Split(db.discovery.Version, ".")[0]) {
		return nil, errors.New("cypher/neohttp: found unsupported version of neo4j - supported versions are: {" +
			strings.Join(supportedMajorVersions, ", ") + "}")
	}
	db.discovery.TX = strings.Replace(db.discovery.TX, "{databaseName}", dbName, 1)
	return db, nil
}

func versionIsSupported(majorVersion string) bool {
	for _, mv := range supportedMajorVersions {
		if mv == majorVersion {
			return true
		}
	}
	return false
}
