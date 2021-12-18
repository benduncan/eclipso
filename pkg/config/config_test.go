package config

import (
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
)

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

	config := Config{}
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

	config := Config{}
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

	config := Config{}
	toml.Unmarshal([]byte(file), &config)
	ApplyDefaults(&config, time.Now())

	//assert.(t, 1.1, config.Version)

	assert.Equal(t, "", config.Domain.Domain)
	assert.Equal(t, 0, len(config.Records))

}
