# dns-go
DNS server using [miekg/dns](https://github.com/miekg/dns) offering dynamic subdomains, time-over-dns, and standard zone file support.

## dynamic subdomains
`web.myapp.192.168.1.13.nip.dns.ribes.ovh.` resolves to A record `192.168.1.13`.
It allows one to use virtualhosts with a local development server

**if local IPs don't work**, disable **DNS rebinding protection** on your network appliances. 

## time over dns
the server will respond to `TXT` and `A` records on `time.some.subdomain.domain.tld` (has to begin with `time`)

## Zone file
create a file `zone.db` in the workding directory. On startup, the file will be read, and upon sucessful parsing, 
will echo back on the command-line.

You can reload the zone with a lookup on `reload-zone.your.subdomain.your.domain.tld`

```dns
$ORIGIN example.com.     ; designates the start of this zone file in the namespace
$TTL 3600                ; default expiration time (in seconds) of all RRs without their own TTL value
@	IN	SOA	localhost. root.localhost. (
			      1		; Serial
			 604800		; Refresh
			  86400		; Retry
			2419200		; Expire
			  86400 )	; Negative Cache TTL
;
@	IN	NS	localhost.
example.com.  IN  SOA   ns.example.com. username.example.com. ( 2020091025 7200 3600 1209600 3600 )
example.com.  IN  NS    ns                    ; ns.example.com is a nameserver for example.com
example.com.  IN  NS    ns.somewhere.example. ; ns.somewhere.example is a backup nameserver for example.com
example.com.  IN  MX    10 mail.example.com.  ; mail.example.com is the mailserver for example.com
@             IN  MX    20 mail2.example.com. ; equivalent to above line, "@" represents zone origin
@             IN  MX    50 mail3              ; equivalent to above line, but using a relative host name
example.com.  IN  A     192.0.2.1             ; IPv4 address for example.com
              IN  AAAA  2001:db8:10::1        ; IPv6 address for example.com
ns            IN  A     192.0.2.2             ; IPv4 address for ns.example.com
              IN  AAAA  2001:db8:10::2        ; IPv6 address for ns.example.com
www           IN  CNAME example.com.          ; www.example.com is an alias for example.com
wwwtest       IN  CNAME www                   ; wwwtest.example.com is another alias for www.example.com
mail          IN  A     192.0.2.3             ; IPv4 address for mail.example.com
mail2         IN  A     192.0.2.4             ; IPv4 address for mail2.example.com
mail3         IN  A     192.0.2.5             ; IPv4 address for mail3.example.com
```

## local test
Use *dig*
```shell
dig @127.0.0.1 -p 8053 foo.bar.1.2.4.4.nip.subdomain.domain.lan
```