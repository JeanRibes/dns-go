package main

import (
	"flag"
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var (
	listen_addr = flag.String("addr", "127.0.0.2:8053", "ip:port")
)

var domainsToAddresses = map[string]string{
	"google.com.":       "1.2.3.4",
	"go.dns.ribes.ovh.": "4.3.2.1",
}
var zone = map[string]map[uint16]dns.RR{}

func serve_dns(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true
	domain := msg.Question[0].Name
	println(" -> ", msg.Question[0].String(), domain)

	if _, zok := zone[domain]; zok { // serve zone file
		if rr_resp, rok := zone[domain][msg.Question[0].Qtype]; rok {
			msg.Answer = []dns.RR{rr_resp}
			w.WriteMsg(&msg)
			return
		}
	}

	start := strings.SplitN(domain, ".", 2)[0]
	if start == "reload-zone" { // reload hook
		msg.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"reloaded"},
		}}
		reload_zone()
		w.WriteMsg(&msg)
		return
	}

	if start == "time" { // return local time
		switch r.Question[0].Qtype {
		case dns.TypeA:
			msg.Answer = []dns.RR{&dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   time_ip(),
			}}
		case dns.TypeTXT:
			msg.Answer = []dns.RR{&dns.TXT{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
				Txt: []string{time.Now().Format("15h04")},
			}}

		}
		w.WriteMsg(&msg)
		return
	}

	switch r.Question[0].Qtype {
	case dns.TypeA:
		address, ok := domainsToAddresses[domain]
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		} else {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   w.RemoteAddr().(*net.UDPAddr).IP,
			})
		}
	case dns.TypeTXT:
		msg.Answer = append(msg.Answer, &dns.TXT{
			Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
			Txt: []string{"hey texte dynamique: " + domain, "lol"},
		})
	}
	w.WriteMsg(&msg)

}

func time_ip() net.IP {
	now := time.Now()
	return net.IPv4(byte(now.Hour()), byte(now.Minute()), byte(now.Second()), 0)
}

func parsezone() {
	f, ferr := os.Open("zone.db")
	if ferr != nil {
		panic(ferr)
	}
	zp := dns.NewZoneParser(f, ".", "zone.db")
	ok := true
	var rr dns.RR
	for ok {
		rr, ok = zp.Next()
		if ok {
			hdr := rr.Header()
			if recs, ok := zone[hdr.Name]; ok {
				recs[hdr.Rrtype] = hdr
			} else {
				zone[hdr.Name] = map[uint16]dns.RR{}
				zone[hdr.Name][hdr.Rrtype] = rr
			}
			println(rr.String())
		}
	}
	if er := zp.Err(); er != nil {
		println(er.Error())
	}
	f.Close()
}

func reload_zone() {
	zone = map[string]map[uint16]dns.RR{}
	parsezone()
}

func main() {
	flag.Parse()
	parsezone()
	dns.HandleFunc(".", serve_dns)
	srv := &dns.Server{Addr: *listen_addr, Net: "udp"}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}

}
