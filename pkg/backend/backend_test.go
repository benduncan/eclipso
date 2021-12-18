package backend_test

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/benduncan/eclipso/pkg/backend"
	"github.com/benduncan/eclipso/pkg/config"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func init() {
	var t *testing.T

	ConfigSetup(t)
	DNSBackend(t)

}
func TestRunner(t *testing.T) {

	//t.Run("A=ConfigSetup", ConfigSetup)
	//t.Run("A=DNSBackend", DNSBackend)
	t.Run("A=LocalQuery", LocalQuery)
	t.Run("A=CleanupConfig", CleanupConfig)

}

func ConfigSetup(t *testing.T) {

	_ = os.Mkdir("./testconfig", 0755)

	//assert.Nil(t, err)

	var md5str string
	iprange := 1

	for i := 'a'; i <= 'z'; i++ {

		md5str = fmt.Sprintf("site-verification-hello_%c.net", i)
		md5str = fmt.Sprintf("%x", md5.Sum([]byte(md5str)))

		filename := fmt.Sprintf("testconfig/hello_%c.net", i)

		output := `
# Domain configration file in TOML format.
version = 1.0

# Domain settings
[domain]
domain = "hello_%c.net"
created = 2021-05-27T07:32:00Z
modified = 2022-05-27T07:32:00Z
verified = true
active = true
ownerid = 10

# Default settings if not defined in each [[records]]
[defaults]
ttl = 3600
type = 1
class = 1

# Domain entry, one entry per record
[[records]]
domain = ""
address = "203.100.%d.1"

[[records]]
domain = "www."
address = "203.100.%d.1"
`

		record := fmt.Sprintf(output, i, iprange, iprange)

		for i2 := 1; i2 < 253; i2++ {

			output = `
[[records]]
domain = "host-%d."
address = "203.100.%d.%d"
		`

			record += fmt.Sprintf(output, i2, iprange, i2)

		}

		var spfips string
		preference := 10

		for i3 := 10; i3 <= 13; i3++ {
			output = `
[[records]]
domain = ""
type = 15
preference = %d
address = "host-%d.hello_%c.net."
					`

			record += fmt.Sprintf(output, preference, i3, i)

			spfips += fmt.Sprintf(" ip:203.100.%d.%d", iprange, i3)
			preference += 10

		}

		output = `
[[records]]
domain = ""
type = 16
address = "v=spf1%s mx a -all"

[[records]]
domain = ""
type = 16
address = "google-site-verification=%s"
		`

		record += fmt.Sprintf(output, spfips, md5str)

		//fmt.Println(record)

		f, err := os.Create(filename)
		if err != nil {
			assert.Nil(t, err)
		}
		defer f.Close()
		if _, err := f.WriteString(record); err != nil {
			assert.Nil(t, err)
		}

		iprange++

	}

}

func DNSBackend(t *testing.T) {

	config.ReadZoneFiles("./testconfig/")

	backend.DomainDB = &config.HostedZones

	var host = "127.0.0.1"
	var port = 5354

	srv := &dns.Server{Addr: fmt.Sprintf("%s:%d", host, port), Net: "udp"}
	srv.Handler = &backend.Handler{}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			assert.Nil(t, err)
		}

	}()

}

func LocalQuery(t *testing.T) {

	ns := "127.0.0.1:5354"
	c := dns.Client{}
	m := dns.Msg{}

	iprange := 1

	for i := 'a'; i <= 'z'; i++ {

		// Step 1: Test each domain resolves
		m.SetQuestion(dns.Fqdn(fmt.Sprintf("hello_%c.net", i)), dns.TypeA)
		r, _, err := c.Exchange(&m, ns)

		if err != nil {
			assert.Nil(t, err)
		}
		if r.Rcode != dns.RcodeSuccess {
			assert.NotEqual(t, r.Rcode, dns.RcodeSuccess)
		}

		for _, k := range r.Answer {
			if key, ok := k.(*dns.A); ok {
				assert.Equal(t, key.A.String(), fmt.Sprintf("203.100.%d.1", iprange))
			}
		}

		//fmt.Println(r)

		// Step 2: Test each subdomain
		for i2 := 1; i2 < 253; i2++ {
			m.SetQuestion(dns.Fqdn(fmt.Sprintf("host-%d.hello_%c.net", i2, i)), dns.TypeA)
			r, _, err := c.Exchange(&m, ns)

			if err != nil {
				assert.Nil(t, err)
			}
			if r.Rcode != dns.RcodeSuccess {
				assert.NotEqual(t, r.Rcode, dns.RcodeSuccess)
			}

			for _, k := range r.Answer {
				if key, ok := k.(*dns.A); ok {
					assert.Equal(t, key.A.String(), fmt.Sprintf("203.100.%d.%d", iprange, i2))
				}
			}

		}

		// Step 3: Test MX records
		m.SetQuestion(dns.Fqdn(fmt.Sprintf("hello_%c.net", i)), dns.TypeMX)
		r, _, err = c.Exchange(&m, ns)

		if err != nil {
			assert.Nil(t, err)
		}
		if r.Rcode != dns.RcodeSuccess {
			assert.NotEqual(t, r.Rcode, dns.RcodeSuccess)
		}

		var mx []string
		var pref []uint16

		for _, k := range r.Answer {
			if key, ok := k.(*dns.MX); ok {
				mx = append(mx, key.Mx)
				pref = append(pref, key.Preference)
			}
		}

		assert.Equal(t, len(mx), 4)
		assert.Equal(t, len(pref), 4)

		if len(mx) > 0 {
			assert.Equal(t, mx[0], fmt.Sprintf("host-10.hello_%c.net.", i))
		}

		// Check MX record preferences are as expected
		if len(pref) == 4 {
			assert.Equal(t, pref[0], uint16(10))
			assert.Equal(t, pref[1], uint16(20))
			assert.Equal(t, pref[2], uint16(30))
			assert.Equal(t, pref[3], uint16(40))
		}

		// Step 4: Test TXT records
		// Step 3: Test MX records
		m.SetQuestion(dns.Fqdn(fmt.Sprintf("hello_%c.net", i)), dns.TypeTXT)
		r, _, err = c.Exchange(&m, ns)

		if err != nil {
			assert.Nil(t, err)
		}
		if r.Rcode != dns.RcodeSuccess {
			assert.NotEqual(t, r.Rcode, dns.RcodeSuccess)
		}

		var txt []string

		for _, k := range r.Answer {
			if key, ok := k.(*dns.TXT); ok {
				txt = append(txt, key.Txt[0])
			}
		}

		assert.Equal(t, 2, len(txt))

		// Check SPF
		spfok := strings.HasPrefix(txt[0], "v=spf1 ip:203")
		assert.Equal(t, spfok, true)

		// Check txt checksum
		md5str := fmt.Sprintf("site-verification-hello_%c.net", i)
		md5str = fmt.Sprintf("%x", md5.Sum([]byte(md5str)))

		assert.Equal(t, txt[1], fmt.Sprintf("google-site-verification=%s", md5str))

		iprange++
	}

	//time.Sleep(time.Minute * 2)
}

func CleanupConfig(t *testing.T) {
	err := os.RemoveAll("./testconfig")
	assert.Nil(t, err)

}

/*
func Fib(n int) int {
	if n < 2 {
		return n
	}
	return Fib(n-1) + Fib(n-2)
}

// from fib_test.go
func BenchmarkFib10(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		Fib(10)
	}
}

*/

func BenchmarkDNSQueryA(b *testing.B) {

	ns := "127.0.0.1:5354"
	c := dns.Client{}
	m := dns.Msg{}

	// Step 1: Test each domain resolves
	for n := 0; n < b.N; n++ {
		m.SetQuestion(dns.Fqdn(fmt.Sprintf("hello_a.net")), dns.TypeA)
		r, _, err := c.Exchange(&m, ns)

		if err != nil {
			//assert.Nil(t, err)
		}
		if r.Rcode != dns.RcodeSuccess {
			//assert.NotEqual(t, r.Rcode, dns.RcodeSuccess)
		}

		for _, k := range r.Answer {
			if _, ok := k.(*dns.A); ok {

				//assert.Equal(t, key.A.String(), fmt.Sprintf("203.100.%d.1", iprange))
			}
		}
	}

}

func BenchmarkDNSQueryTXT(b *testing.B) {

	ns := "127.0.0.1:5354"
	c := dns.Client{}
	m := dns.Msg{}

	// Step 1: Test each domain resolves
	for n := 0; n < b.N; n++ {
		m.SetQuestion(dns.Fqdn(fmt.Sprintf("hello_a.net")), dns.TypeTXT)
		r, _, err := c.Exchange(&m, ns)

		if err != nil {
			//assert.Nil(t, err)
		}
		if r.Rcode != dns.RcodeSuccess {
			//assert.NotEqual(t, r.Rcode, dns.RcodeSuccess)
		}

		for _, k := range r.Answer {
			if _, ok := k.(*dns.TXT); ok {
				//assert.Equal(t, key.A.String(), fmt.Sprintf("203.100.%d.1", iprange))
			}
		}
	}

}

func BenchmarkDNSQueryMX(b *testing.B) {

	ns := "127.0.0.1:5354"
	c := dns.Client{}
	m := dns.Msg{}

	// Step 1: Test each domain resolves
	for n := 0; n < b.N; n++ {
		m.SetQuestion(dns.Fqdn(fmt.Sprintf("hello_a.net")), dns.TypeMX)
		r, _, err := c.Exchange(&m, ns)

		if err != nil {
			//assert.Nil(t, err)
		}
		if r.Rcode != dns.RcodeSuccess {
			//assert.NotEqual(t, r.Rcode, dns.RcodeSuccess)
		}

		for _, k := range r.Answer {
			if _, ok := k.(*dns.MX); ok {
				//assert.Equal(t, key.A.String(), fmt.Sprintf("203.100.%d.1", iprange))
			}
		}
	}

}
