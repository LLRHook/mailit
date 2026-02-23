package engine

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// MXRecord represents an MX DNS record with its host and priority.
type MXRecord struct {
	Host     string
	Priority uint16
}

// DNSRecord represents a DNS record that needs to be configured for email delivery.
type DNSRecord struct {
	RecordType string // SPF, DKIM, MX, DMARC, RETURN_PATH
	DNSType    string // TXT, MX, CNAME
	Name       string
	Value      string
	Priority   *int
}

// DNSResolver performs DNS lookups. It can be configured to use a specific
// nameserver or fall back to the system resolver.
type DNSResolver struct {
	nameserver string
	timeout    time.Duration
}

// NewDNSResolver creates a new DNS resolver. If nameserver is empty or "system",
// it uses the system's default resolver (8.8.8.8:53 as fallback).
func NewDNSResolver(nameserver string, timeout time.Duration) *DNSResolver {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	if nameserver == "" || nameserver == "system" {
		nameserver = getSystemResolver()
	}
	if !strings.Contains(nameserver, ":") {
		nameserver = nameserver + ":53"
	}
	return &DNSResolver{
		nameserver: nameserver,
		timeout:    timeout,
	}
}

// getSystemResolver attempts to read the system's DNS resolver. Falls back to
// Google Public DNS if detection fails.
func getSystemResolver() string {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err == nil && len(config.Servers) > 0 {
		return config.Servers[0] + ":53"
	}
	return "8.8.8.8:53"
}

// query performs a DNS query for the given name and type.
func (r *DNSResolver) query(name string, qtype uint16) (*dns.Msg, error) {
	c := &dns.Client{
		Timeout: r.timeout,
	}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), qtype)
	m.RecursionDesired = true

	reply, _, err := c.Exchange(m, r.nameserver)
	if err != nil {
		return nil, fmt.Errorf("DNS query for %s (type %s): %w", name, dns.TypeToString[qtype], err)
	}
	if reply.Rcode != dns.RcodeSuccess {
		return reply, fmt.Errorf("DNS query for %s returned %s", name, dns.RcodeToString[reply.Rcode])
	}

	return reply, nil
}

// LookupMX resolves MX records for a domain, sorted by priority (lowest first).
func (r *DNSResolver) LookupMX(domain string) ([]MXRecord, error) {
	reply, err := r.query(domain, dns.TypeMX)
	if err != nil {
		return nil, fmt.Errorf("looking up MX for %s: %w", domain, err)
	}

	var records []MXRecord
	for _, ans := range reply.Answer {
		if mx, ok := ans.(*dns.MX); ok {
			records = append(records, MXRecord{
				Host:     strings.TrimSuffix(mx.Mx, "."),
				Priority: mx.Preference,
			})
		}
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Priority < records[j].Priority
	})

	// If no MX records found, fall back to the domain's A/AAAA record per RFC 5321.
	if len(records) == 0 {
		records = append(records, MXRecord{
			Host:     domain,
			Priority: 0,
		})
	}

	return records, nil
}

// lookupTXT fetches all TXT records for a name and returns them joined.
func (r *DNSResolver) lookupTXT(name string) ([]string, error) {
	reply, err := r.query(name, dns.TypeTXT)
	if err != nil {
		return nil, err
	}

	var records []string
	for _, ans := range reply.Answer {
		if txt, ok := ans.(*dns.TXT); ok {
			// TXT records can be split across multiple strings; join them.
			records = append(records, strings.Join(txt.Txt, ""))
		}
	}

	return records, nil
}

// VerifySPF checks if the domain has a valid SPF record that includes the
// expected hostname. Returns whether the record exists, the full SPF value,
// and any error.
func (r *DNSResolver) VerifySPF(domain string, expectedHostname string) (bool, string, error) {
	records, err := r.lookupTXT(domain)
	if err != nil {
		return false, "", fmt.Errorf("looking up SPF for %s: %w", domain, err)
	}

	for _, record := range records {
		if strings.HasPrefix(record, "v=spf1") {
			if expectedHostname == "" {
				return true, record, nil
			}
			if strings.Contains(record, "include:"+expectedHostname) ||
				strings.Contains(record, "a:"+expectedHostname) ||
				strings.Contains(record, "mx:"+expectedHostname) {
				return true, record, nil
			}
			return false, record, nil
		}
	}

	return false, "", nil
}

// VerifyDKIM checks if the DKIM TXT record exists at selector._domainkey.domain.
// Returns whether the record exists, the full record value, and any error.
func (r *DNSResolver) VerifyDKIM(domain, selector string) (bool, string, error) {
	name := selector + "._domainkey." + domain
	records, err := r.lookupTXT(name)
	if err != nil {
		// NXDOMAIN is not an error here, it means the record doesn't exist.
		return false, "", nil
	}

	for _, record := range records {
		if strings.Contains(record, "v=DKIM1") || strings.Contains(record, "p=") {
			return true, record, nil
		}
	}

	return false, "", nil
}

// VerifyMX checks if MX records for the domain include the expected host.
func (r *DNSResolver) VerifyMX(domain, expectedHost string) (bool, error) {
	records, err := r.LookupMX(domain)
	if err != nil {
		return false, fmt.Errorf("looking up MX for %s: %w", domain, err)
	}

	expectedHost = strings.TrimSuffix(strings.ToLower(expectedHost), ".")
	for _, mx := range records {
		if strings.ToLower(mx.Host) == expectedHost {
			return true, nil
		}
	}

	return false, nil
}

// VerifyDMARC checks if the domain has a DMARC record at _dmarc.domain.
// Returns whether the record exists, the full record value, and any error.
func (r *DNSResolver) VerifyDMARC(domain string) (bool, string, error) {
	name := "_dmarc." + domain
	records, err := r.lookupTXT(name)
	if err != nil {
		return false, "", nil
	}

	for _, record := range records {
		if strings.HasPrefix(record, "v=DMARC1") {
			return true, record, nil
		}
	}

	return false, "", nil
}

// VerifyReturnPath checks if the bounce subdomain CNAME points to the expected host.
func (r *DNSResolver) VerifyReturnPath(domain, expectedHost string) (bool, error) {
	name := "bounce." + domain
	reply, err := r.query(name, dns.TypeCNAME)
	if err != nil {
		return false, nil
	}

	expectedHost = strings.TrimSuffix(strings.ToLower(expectedHost), ".")
	for _, ans := range reply.Answer {
		if cname, ok := ans.(*dns.CNAME); ok {
			target := strings.TrimSuffix(strings.ToLower(cname.Target), ".")
			if target == expectedHost {
				return true, nil
			}
		}
	}

	return false, nil
}

// ResolveIP resolves an MX host to its IP addresses for SMTP connection.
func (r *DNSResolver) ResolveIP(host string) ([]net.IP, error) {
	var ips []net.IP

	// Try A records first.
	replyA, err := r.query(host, dns.TypeA)
	if err == nil {
		for _, ans := range replyA.Answer {
			if a, ok := ans.(*dns.A); ok {
				ips = append(ips, a.A)
			}
		}
	}

	// Also try AAAA records.
	replyAAAA, err := r.query(host, dns.TypeAAAA)
	if err == nil {
		for _, ans := range replyAAAA.Answer {
			if aaaa, ok := ans.(*dns.AAAA); ok {
				ips = append(ips, aaaa.AAAA)
			}
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no A or AAAA records found for %s", host)
	}

	return ips, nil
}

// GenerateDNSRecords generates the complete set of DNS records that a user must
// configure for their domain to send email through MailIt.
func GenerateDNSRecords(domain, selector, dkimPublicKey, hostname string) []DNSRecord {
	mxPriority := 10

	records := []DNSRecord{
		{
			RecordType: "SPF",
			DNSType:    "TXT",
			Name:       domain,
			Value:      fmt.Sprintf("v=spf1 include:%s ~all", hostname),
		},
		{
			RecordType: "DKIM",
			DNSType:    "TXT",
			Name:       fmt.Sprintf("%s._domainkey.%s", selector, domain),
			Value:      fmt.Sprintf("v=DKIM1; k=rsa; p=%s", dkimPublicKey),
		},
		{
			RecordType: "MX",
			DNSType:    "MX",
			Name:       domain,
			Value:      hostname,
			Priority:   &mxPriority,
		},
		{
			RecordType: "DMARC",
			DNSType:    "TXT",
			Name:       fmt.Sprintf("_dmarc.%s", domain),
			Value:      fmt.Sprintf("v=DMARC1; p=none; rua=mailto:dmarc@%s", domain),
		},
		{
			RecordType: "RETURN_PATH",
			DNSType:    "CNAME",
			Name:       fmt.Sprintf("bounce.%s", domain),
			Value:      hostname,
		},
	}

	return records
}
