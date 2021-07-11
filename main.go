package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type Domains struct {
	Domain  string
	Type    uint16
	Class   uint16
	TTL     uint32
	Address string
}

var domains []Domains

type handler struct{}

func (this *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	fmt.Println(r)

	domain := msg.Question[0].Name

	fmt.Println(domain)

	log.Printf("DNS Request: %q => %q (type %d)", domain, w.RemoteAddr(), r.Question[0].Qtype)

	records, ok := lookupDomain(domain)

	if !ok {
		// Return `NODATA` response
		msg.SetRcode(r, dns.RcodeRefused)
		w.WriteMsg(&msg)
		// Consider delay/rate limit to prevent abuse
		return

	}

	// Loop through the domain records and append a response for each
	for i := 0; i < len(records); i++ {
		record := records[i]

		// The domain is authoritative
		msg.Authoritative = true

		fmt.Println("TYPE => ", r.Question[0].Qtype)

		// Case type, switch for each supported type
		switch r.Question[0].Qtype {

		// TODO: AAAA records

		// A records
		case dns.TypeA:

			if record.Type == dns.TypeA {
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					A:   net.ParseIP(record.Address),
				})
			}

			if record.Type == dns.TypeCNAME {
				msg.Answer = append(msg.Answer, &dns.CNAME{
					Hdr:    dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					Target: record.Address,
				})

				lookupRecord, err := lookupHost(record.Address, 3)

				fmt.Println(lookupRecord)

				if err == nil {

					for _, record := range lookupRecord {
						if t, ok := record.(*dns.A); ok {

							msg.Answer = append(msg.Answer, &dns.A{
								Hdr: dns.RR_Header{Name: t.Hdr.Name, Rrtype: t.Hdr.Rrtype, Class: t.Hdr.Class, Ttl: t.Hdr.Ttl},
								A:   t.A,
							})

						}
					}

				}

				// Next, fetch the A record

			}

		// CNAME case -- Confirm syntax
		case dns.TypeCNAME:

			if record.Type == dns.TypeCNAME {
				msg.Answer = append(msg.Answer, &dns.CNAME{
					Hdr:    dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					Target: record.Address,
				})
			}

		// SOA case
		case dns.TypeSOA:
			msg.Answer = append(msg.Answer, SOA(record.Domain))
			// TODO: Only return one

		// NS case
		case dns.TypeNS:
			fmt.Println("NS TYPE")
			if record.Type == dns.TypeNS {

				msg.Answer = append(msg.Answer, &dns.NS{
					Hdr: dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					Ns:  record.Address,
				})
				// TODO
				// Lookup all A, AAAA records, e.g ns1.domain.com > 123.43.14.1

			}

		default:
			fmt.Println(r.Question[0].Qtype)
			msg.SetRcode(r, dns.RcodeRefused)

		}

	}

	// Return `NXDOMAIN`, we are authoritative however domain does not exist.
	if len(msg.Answer) == 0 {
		msg.SetRcode(r, dns.RcodeNameError)
		msg.Ns = []dns.RR{SOA(domain)}
	}

	w.WriteMsg(&msg)

}

func lookupDomain(domain string) (d []Domains, s bool) {

	for i := 0; i < len(domains); i++ {

		if domains[i].Domain == domain {
			d = append(d, domains[i])
		}

	}

	if len(d) > 0 {
		s = true
	}

	return d, s
}

func lookupHost(host string, triesLeft int) ([]dns.RR, error) {
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{dns.Fqdn(host), dns.TypeA, dns.ClassINET}
	in, err := dns.Exchange(m1, "1.1.1.1:53")

	if err != nil {
		if strings.HasSuffix(err.Error(), "i/o timeout") && triesLeft > 0 {
			triesLeft--
			return lookupHost(host, triesLeft)
		}
		return nil, err
	}

	if in != nil && in.Rcode != dns.RcodeSuccess {
		return nil, errors.New(dns.RcodeToString[in.Rcode])
	}

	return in.Answer, err
}

func main() {
	domains = make([]Domains, 16)

	// TODO: Move to a JSON file, obj/storage, NoSQL
	domains[0].Domain = "test.com."
	domains[0].Type = dns.TypeA
	domains[0].Class = dns.ClassINET
	domains[0].TTL = 300
	domains[0].Address = "1.1.1.1"

	domains[1].Domain = "google.com."
	domains[1].Type = dns.TypeA
	domains[1].Class = dns.ClassINET
	domains[1].TTL = 300
	domains[1].Address = "8.8.8.8"

	domains[2].Domain = "m.facebook.com."
	domains[2].Type = dns.TypeCNAME
	domains[2].Class = dns.ClassINET
	domains[2].TTL = 300
	domains[2].Address = "star-mini.c10r.facebook.com."

	domains[3].Domain = "xeon.us-west-1.phasegrid.net."
	domains[3].Type = dns.TypeA
	domains[3].Class = dns.ClassINET
	domains[3].TTL = 300
	domains[3].Address = "216.218.163.99"

	domains[4].Domain = "ilo.xeon.us-west-1.phasegrid.net."
	domains[4].Type = dns.TypeA
	domains[4].Class = dns.ClassINET
	domains[4].TTL = 60
	domains[4].Address = "216.218.163.98"

	domains[5].Domain = "phasegrid.net."
	domains[5].Type = dns.TypeNS
	domains[5].Class = dns.ClassINET
	domains[5].TTL = 60 * 60 * 24
	domains[5].Address = "ns1.phasegrid.net."

	domains[6].Domain = "phasegrid.net."
	domains[6].Type = dns.TypeNS
	domains[6].Class = dns.ClassINET
	domains[6].TTL = 60 * 60 * 24
	domains[6].Address = "ns2.phasegrid.net."

	domains[7].Domain = "phasegrid.net."
	domains[7].Type = dns.TypeNS
	domains[7].Class = dns.ClassINET
	domains[7].TTL = 60 * 60 * 24
	domains[7].Address = "ns3.phasegrid.net."

	domains[8].Domain = "ns1.phasegrid.net."
	domains[8].Type = dns.TypeA
	domains[8].Class = dns.ClassINET
	domains[8].TTL = 60 * 60 * 24
	domains[8].Address = "216.218.163.102"

	domains[9].Domain = "ns2.phasegrid.net."
	domains[9].Type = dns.TypeA
	domains[9].Class = dns.ClassINET
	domains[9].TTL = 60 * 60 * 24
	domains[9].Address = "216.218.163.101"

	domains[10].Domain = "ns3.phasegrid.net."
	domains[10].Type = dns.TypeA
	domains[10].Class = dns.ClassINET
	domains[10].TTL = 60 * 60 * 24
	domains[10].Address = "129.159.43.166"

	domains[11].Domain = "neon.us-west-2.phasegrid.net."
	domains[11].Type = dns.TypeA
	domains[11].Class = dns.ClassINET
	domains[11].TTL = 60 * 60 * 24
	domains[11].Address = "152.67.248.9"

	domains[12].Domain = "radon.us-west-1.phasegrid.net."
	domains[12].Type = dns.TypeA
	domains[12].Class = dns.ClassINET
	domains[12].TTL = 300
	domains[12].Address = "216.218.163.104"

	domains[13].Domain = "idrac.radon.us-west-1.phasegrid.net."
	domains[13].Type = dns.TypeA
	domains[13].Class = dns.ClassINET
	domains[13].TTL = 300
	domains[13].Address = "216.218.163.100"

	domains[14].Domain = "pfsense.radon.us-west-1.phasegrid.net."
	domains[14].Type = dns.TypeA
	domains[14].Class = dns.ClassINET
	domains[14].TTL = 300
	domains[14].Address = "216.218.163.105"

	domains[15].Domain = "pfsense.xeon.us-west-1.phasegrid.net."
	domains[15].Type = dns.TypeA
	domains[15].Class = dns.ClassINET
	domains[15].TTL = 300
	domains[15].Address = "216.218.163.103"

	// 216.218.163.99

	var host = os.Getenv("HOST")

	if host == "" {
		host = "0.0.0.0"
	}

	srv := &dns.Server{Addr: host + ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}

func SOA(domain string) dns.RR {
	return &dns.SOA{Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60},
		Ns:      "master." + domain,
		Mbox:    "hostmaster." + domain,
		Serial:  uint32(time.Now().Truncate(time.Hour).Unix()),
		Refresh: 28800,
		Retry:   7200,
		Expire:  604800,
		Minttl:  60,
	}
}
