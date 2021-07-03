package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

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

	record, ok := lookupDomain(domain)

	if !ok {
		return
	}

	// Case type
	switch r.Question[0].Qtype {

	// A records
	case dns.TypeA:
		msg.Authoritative = true

		if ok && record.Type == dns.TypeA {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
				A:   net.ParseIP(record.Address),
			})
		}

		if ok && record.Type == dns.TypeCNAME {
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

						//result = append(result, t.A)

					}
				}

				/*
					msg.Answer = append(msg.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: record.Domain, Rrtype: dns.TypeA, Class: record.Class, Ttl: record.TTL},
						A:   lookupRecord[0],
					})
				*/

			}

			// Next, fetch the A record

		}

	case dns.TypeCNAME:

		if ok {

			msg.Answer = append(msg.Answer, &dns.CNAME{
				Hdr:    dns.RR_Header{Name: record.Domain, Rrtype: record.Type, Class: record.Class, Ttl: record.TTL},
				Target: record.Address,
			})

			//fmt.Println(lastCNAME(record.Address))

			//msg.Answer = append(msg.Answer, &dns.CNAME{
			//	Target: record.Address,
			//})

		}

	}

	w.WriteMsg(&msg)

}

func lookupDomain(domain string) (d Domains, s bool) {

	for i := 0; i < len(domains); i++ {

		if domains[i].Domain == domain {
			return domains[i], true
		}

	}

	return d, false
}

func lookupHost(host string, triesLeft int) ([]dns.RR, error) {
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{dns.Fqdn(host), dns.TypeA, dns.ClassINET}
	in, err := dns.Exchange(m1, "1.1.1.1:53")

	//result := []net.IP{}

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

	/*
		for _, record := range in.Answer {
			if t, ok := record.(*dns.A); ok {
				result = append(result, t.A)
			}
		}
	*/

	return in.Answer, err
}

func main() {
	domains = make([]Domains, 3)

	domains[0].Domain = "test.com."
	domains[0].Type = dns.TypeA
	domains[0].Class = dns.ClassINET
	domains[0].TTL = 60
	domains[0].Address = "1.1.1.1"

	domains[1].Domain = "google.com."
	domains[1].Type = dns.TypeA
	domains[1].Class = dns.ClassINET
	domains[1].TTL = 60
	domains[1].Address = "8.8.8.8"

	domains[2].Domain = "m.facebook.com."
	domains[2].Type = dns.TypeCNAME
	domains[2].Class = dns.ClassINET
	domains[2].TTL = 60
	domains[2].Address = "star-mini.c10r.facebook.com."

	domains[3].Domain = "xeon.us-west-1.phasegrid.net."
	domains[3].Type = dns.TypeA
	domains[3].Class = dns.ClassINET
	domains[3].TTL = 60
	domains[3].Address = "216.218.163.99"

	domains[4].Domain = "xeon-ilo.us-west-1.phasegrid.net."
	domains[4].Type = dns.TypeA
	domains[4].Class = dns.ClassINET
	domains[4].TTL = 60
	domains[4].Address = "216.218.163.98"

	// 216.218.163.99

	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
