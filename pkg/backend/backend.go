package backend

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/benduncan/eclipso/pkg/config"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

var DomainDB *[]config.Config

type Handler struct{}

func (this *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	domain := msg.Question[0].Name

	_, logignore := os.LookupEnv("ECLIPSO_LOG_IGNORE")

	if logignore {
		log.SetLevel(log.FatalLevel)
	}

	log.Printf("DNS Request: %q => %q (type %d)", domain, w.RemoteAddr(), r.Question[0].Qtype)

	// Return matching domain records for the request
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

		// Case type, switch for each supported type
		switch r.Question[0].Qtype {

		// AAAA records
		case dns.TypeAAAA:

			if record.Type == dns.TypeAAAA {
				msg.Answer = append(msg.Answer, &dns.AAAA{
					Hdr:  dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					AAAA: net.ParseIP(record.Address),
				})
			}

		// A records
		case dns.TypeA:

			if record.Type == dns.TypeA {
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					A:   net.ParseIP(record.Address),
				})
			}

			// CONFIRM CORRECT. CNAME > external lookup?!
			if record.Type == dns.TypeCNAME {
				msg.Answer = append(msg.Answer, &dns.CNAME{
					Hdr:    dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					Target: record.Address,
				})

				lookupRecord, err := lookupHost(record.Address, 3)

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
			msg.Answer = []dns.RR{SOA(record.Domain)}
			continue

		// TXT record case
		case dns.TypeTXT:

			var txt []string

			if record.Type == dns.TypeTXT {
				txt = append(txt, record.Address)

				msg.Answer = append(msg.Answer, &dns.TXT{
					Hdr: dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					Txt: txt,
				})
			}

		// MX record case
		case dns.TypeMX:

			if record.Type == dns.TypeMX {

				msg.Answer = append(msg.Answer, &dns.MX{
					Hdr:        dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					Preference: record.Preference,
					Mx:         record.Address,
				})

				// Lookup additional A/AAAA records
				extra := lookupExtra(record.Address)
				if extra != nil {
					for _, rr := range extra {
						msg.Extra = append(msg.Extra, rr)
					}
				}

			}

		// NS case
		case dns.TypeNS:

			if record.Type == dns.TypeNS {

				msg.Answer = append(msg.Answer, &dns.NS{
					Hdr: dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
					Ns:  record.Address,
				})

				// Lookup additional A/AAAA records
				extra := lookupExtra(record.Address)
				if extra != nil {
					for _, rr := range extra {
						msg.Extra = append(msg.Extra, rr)
					}
				}

			}

		default:
			msg.SetRcode(r, dns.RcodeRefused)

		}

	}

	// Return `NXDOMAIN`, we are authoritative however domain does not exist.
	if len(msg.Answer) == 0 {
		msg.SetRcode(r, dns.RcodeNameError)
		msg.Ns = []dns.RR{SOA(domain)}
	}

	// Check max record length
	if msg.Len() > 9192 {
		log.Warn(msg.Answer[0].Header().Name, "=> Response too large, returning error")
		msg.Answer = nil
		msg.SetRcode(r, dns.RcodeServerFailure)
	}

	w.WriteMsg(&msg)

}

func lookupExtra(address string) (msg []dns.RR) {

	extraRecords, _ := lookupDomain(address)

	// Append extra fields if required
	for i2 := 0; i2 < len(extraRecords); i2++ {
		r := extraRecords[i2]

		if r.Domain == address && r.Type == dns.TypeA {
			msg = append(msg, &dns.A{
				Hdr: dns.RR_Header{Name: r.Domain, Rrtype: r.Type, Class: r.Class, Ttl: r.TTL},
				A:   net.ParseIP(r.Address),
			})
		}

		if r.Domain == address && r.Type == dns.TypeAAAA {
			msg = append(msg, &dns.AAAA{
				Hdr:  dns.RR_Header{Name: r.Domain, Rrtype: r.Type, Class: r.Class, Ttl: r.TTL},
				AAAA: net.ParseIP(r.Address),
			})

		}

	}

	return msg

}

// TODO: Improve the lookup function, use a pointer?
func lookupDomain(domain string) (d []config.Records, s bool) {

	for i := 0; i < len((*DomainDB)); i++ {

		for i2 := 0; i2 < len((*DomainDB)[i].Records); i2++ {

			if (*DomainDB)[i].Records[i2].Domain == domain {
				d = append(d, (*DomainDB)[i].Records[i2])
			}

		}

	}

	if len(d) > 0 {
		s = true
	}

	return d, s
}

// TODO: Improve the lookup function, use a pointer?
func lookupSOA(domain string) (soa string) {

	for i := 0; i < len((*DomainDB)); i++ {

		for i2 := 0; i2 < len((*DomainDB)[i].Records); i2++ {

			if (*DomainDB)[i].Records[i2].Domain == domain {
				soa = (*DomainDB)[i].Domain.SOA
				break
			}

		}

	}

	// If no SOA exists, return a record (however invalid, need to add improved checks)
	if soa == "" {
		soa = fmt.Sprintf("ns.%s", domain)
	}

	return
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

func SOA(domain string) dns.RR {
	return &dns.SOA{Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60},
		Ns:      lookupSOA(domain),
		Mbox:    "hostmaster." + domain,
		Serial:  uint32(time.Now().Truncate(time.Hour).Unix()),
		Refresh: 28800,
		Retry:   7200,
		Expire:  604800,
		Minttl:  60,
	}
}
