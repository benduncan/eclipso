package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
)

func TestDomainLookup(t *testing.T) {
	conf := GenerateTestDomains(1000)

	assert.Equal(t, 1000, len(conf.Domain))

	// Lookup a domain easily
	for i := 0; i < 1000; i++ {
		lookup := DomainLookup{Domain: fmt.Sprintf("test%d.net", i), Type: 1, Class: 1}

		assert.Equal(t, 4, len(conf.Records[lookup]))

		for i2 := 1; i < 5; i++ {

			if len(conf.Records[lookup]) == 4 {
				assert.Equal(t, fmt.Sprintf("213.189.1.%d", i2), conf.Records[lookup][i2-1].Address)
			}

		}

	}

	// Find records for test1.net
	/*
		records := conf.Domain["test1.net"].RecordRef

		// Delete marked domains
		for _, v := range records {
			delete(conf.Records, v)
		}

		if entry, ok := conf.Domain["test1.net"]; ok {
			entry.RecordRef = []DomainLookup{}
			conf.Domain["test1.net"] = entry
		}

		fmt.Println("Post rm Record =>", conf.Records[records[0]])
		fmt.Println("Post RecordRef =>", conf.Domain["test1.net"].RecordRef)
	*/

	fmt.Println("RecordRef =>", conf.Domain["test1.net"].RecordRef)
	records := conf.Domain["test1.net"].RecordRef

	conf.DeleteZone("test1.net")
	fmt.Println("Post rm Record =>", conf.Records[records[0]])
	fmt.Println("Post RecordRef =>", conf.Domain["test1.net"].RecordRef)

}

func TestConfigGood(t *testing.T) {

	file := `
# This is a TOML document. Boom.
version = 1.1

[domain]
domain = "hotastest.net"
created = 2021-05-27T07:32:00Z
modified = 2022-05-27T07:32:00Z
verified = true
active = true
ownerid = 10

[defaults]
ttl = 3600
type = 1
class = 1

[[records]]
domain = "web3.defi."
address = "213.189.1.4"

[[records]]
domain = "www."
type = 2
class = 1
address = "e15316.a.akamaiedge.net."
`

	config := ConfigArr{}
	toml.Unmarshal([]byte(file), &config)
	ApplyDefaults(&config, time.Now())

	//assert.(t, 1.1, config.Version)

	assert.Equal(t, "hotastest.net", config.Domain.Domain)

	assert.Equal(t, "web3.defi.hotastest.net.", config.Records[0].Domain)
	assert.Equal(t, "213.189.1.4", config.Records[0].Address)
	assert.Equal(t, uint16(1), config.Records[0].Type)
	assert.Equal(t, uint16(1), config.Records[0].Class)
	assert.Equal(t, uint32(3600), config.Records[0].TTL)

	assert.Equal(t, "www.hotastest.net.", config.Records[1].Domain)
	assert.Equal(t, uint16(2), config.Records[1].Type)
	assert.Equal(t, uint16(1), config.Records[1].Class)
	assert.Equal(t, "e15316.a.akamaiedge.net.", config.Records[1].Address)

}

func TestConfigDefaults(t *testing.T) {

	file := `
# This is a TOML document. Boom.
version = 1.1

[domain]
domain = "nodefaults.net"
verified = true
active = true

[[records]]
domain = "web3.defi."
address = "213.189.1.4"
`

	config := ConfigArr{}
	toml.Unmarshal([]byte(file), &config)
	ApplyDefaults(&config, time.Now())

	//assert.(t, 1.1, config.Version)

	assert.Equal(t, "nodefaults.net", config.Domain.Domain)
	assert.Equal(t, uint16(1), config.Records[0].Type)
	assert.Equal(t, uint16(1), config.Records[0].Class)
	assert.Equal(t, uint32(3600), config.Records[0].TTL)

}

func TestConfigBad(t *testing.T) {
	file := `
# This is a TOML document. Boom.
versionz = 1.1

[domainz]
domain = "bad.domain.net"

[[norecords]]
domain = "bad.defi."

[[norecords]]
domain = "www."
`

	config := ConfigArr{}
	toml.Unmarshal([]byte(file), &config)
	ApplyDefaults(&config, time.Now())

	//assert.(t, 1.1, config.Version)

	assert.Equal(t, "", config.Domain.Domain)
	assert.Equal(t, 0, len(config.Records))

}
