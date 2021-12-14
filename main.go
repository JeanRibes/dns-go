package main

import (
	"flag"
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	listen_addr = flag.String("addr", "0.0.0.0:8053", "ip:port")
	zone_file   = flag.String("zone", "zone.db", "zone file")
)

var zone = map[string]map[uint16][]dns.RR{}

func serve_dns(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true
	domain := msg.Question[0].Name
	println(" -> ", msg.Question[0].String(), domain)

	if _, zok := zone[domain]; zok { // serve zone file
		if rr_resp, rok := zone[domain][msg.Question[0].Qtype]; rok {
			msg.Answer = rr_resp
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
		// sub.domains.192.168.1.1.xip.sub.domain.tld returns 192.168.1.1
		parts := strings.Split(domain, ".")
		xindex := -1
		for i := 0; i < len(parts); i++ {
			if parts[i] == "nip" {
				xindex = i
			}
		}
		if xindex-4 >= 0 {
			ip0, er0 := strconv.ParseInt(parts[xindex-1], 10, 16)
			ip1, er1 := strconv.ParseInt(parts[xindex-2], 10, 16)
			ip2, er2 := strconv.ParseInt(parts[xindex-3], 10, 16)
			ip3, er3 := strconv.ParseInt(parts[xindex-4], 10, 16)
			if er0 != nil || er1 != nil || er2 != nil || er3 != nil {
				msg.SetRcode(r, dns.RcodeServerFailure)
			}
			ip := net.IPv4(byte(ip3), byte(ip2), byte(ip1), byte(ip0))

			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   ip,
			})
		} else {
			msg.SetRcode(r, dns.RcodeServerFailure)
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
	f, ferr := os.Open(*zone_file)
	if ferr != nil {
		println("failed to open file", ferr)
		return
	} else {
		defer f.Close()
	}
	zp := dns.NewZoneParser(f, ".", "zone.db")
	//zp := dns.NewZoneParser(f, ".", *zone_file)
	ok := true
	var rr dns.RR
	for ok {
		rr, ok = zp.Next()
		if ok {
			hdr := rr.Header()
			if _, ok := zone[hdr.Name]; ok {
				zone[hdr.Name][hdr.Rrtype] = append(zone[hdr.Name][hdr.Rrtype], rr)
			} else {
				zone[hdr.Name] = map[uint16][]dns.RR{}
				zone[hdr.Name][hdr.Rrtype] = []dns.RR{rr}
			}
			println(rr.String())
		}
	}
	if er := zp.Err(); er != nil {
		println(er.Error())
	}
}

func reload_zone() {
	defer func() { recover() }()
	zone = map[string]map[uint16][]dns.RR{}
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
