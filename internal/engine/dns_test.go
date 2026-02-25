package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDNSRecords(t *testing.T) {
	domain := "example.com"
	selector := "mailit"
	dkimPublicKey := "MIIBIjANBgkqhki..."
	hostname := "mail.example.com"

	records := GenerateDNSRecords(domain, selector, dkimPublicKey, hostname)

	t.Run("returns exactly 5 records", func(t *testing.T) {
		require.Len(t, records, 5)
	})

	t.Run("SPF record", func(t *testing.T) {
		spf := records[0]
		assert.Equal(t, "SPF", spf.RecordType)
		assert.Equal(t, "TXT", spf.DNSType)
		assert.Equal(t, domain, spf.Name)
		assert.Equal(t, fmt.Sprintf("v=spf1 include:%s ~all", hostname), spf.Value)
		assert.Nil(t, spf.Priority)
	})

	t.Run("DKIM record", func(t *testing.T) {
		dkim := records[1]
		assert.Equal(t, "DKIM", dkim.RecordType)
		assert.Equal(t, "TXT", dkim.DNSType)
		assert.Equal(t, fmt.Sprintf("%s._domainkey.%s", selector, domain), dkim.Name)
		assert.Equal(t, fmt.Sprintf("v=DKIM1; k=rsa; p=%s", dkimPublicKey), dkim.Value)
		assert.Nil(t, dkim.Priority)
	})

	t.Run("MX record", func(t *testing.T) {
		mx := records[2]
		assert.Equal(t, "MX", mx.RecordType)
		assert.Equal(t, "MX", mx.DNSType)
		assert.Equal(t, domain, mx.Name)
		assert.Equal(t, hostname, mx.Value)
		require.NotNil(t, mx.Priority)
		assert.Equal(t, 10, *mx.Priority)
	})

	t.Run("DMARC record", func(t *testing.T) {
		dmarc := records[3]
		assert.Equal(t, "DMARC", dmarc.RecordType)
		assert.Equal(t, "TXT", dmarc.DNSType)
		assert.Equal(t, fmt.Sprintf("_dmarc.%s", domain), dmarc.Name)
		assert.Equal(t, fmt.Sprintf("v=DMARC1; p=none; rua=mailto:dmarc@%s", domain), dmarc.Value)
		assert.Nil(t, dmarc.Priority)
	})

	t.Run("RETURN_PATH record", func(t *testing.T) {
		rp := records[4]
		assert.Equal(t, "RETURN_PATH", rp.RecordType)
		assert.Equal(t, "CNAME", rp.DNSType)
		assert.Equal(t, fmt.Sprintf("bounce.%s", domain), rp.Name)
		assert.Equal(t, hostname, rp.Value)
		assert.Nil(t, rp.Priority)
	})

	t.Run("only MX has priority set", func(t *testing.T) {
		for i, r := range records {
			if r.RecordType == "MX" {
				assert.NotNil(t, r.Priority, "MX record should have priority")
			} else {
				assert.Nil(t, r.Priority, "record %d (%s) should not have priority", i, r.RecordType)
			}
		}
	})
}

func TestGenerateDNSRecords_DifferentInputs(t *testing.T) {
	t.Run("different domain and hostname", func(t *testing.T) {
		records := GenerateDNSRecords("myapp.io", "selector1", "pubkey123", "smtp.myapp.io")
		require.Len(t, records, 5)

		assert.Equal(t, "myapp.io", records[0].Name) // SPF
		assert.Contains(t, records[0].Value, "include:smtp.myapp.io")
		assert.Equal(t, "selector1._domainkey.myapp.io", records[1].Name) // DKIM
		assert.Contains(t, records[1].Value, "p=pubkey123")
		assert.Equal(t, "_dmarc.myapp.io", records[3].Name) // DMARC
		assert.Contains(t, records[3].Value, "mailto:dmarc@myapp.io")
		assert.Equal(t, "bounce.myapp.io", records[4].Name) // RETURN_PATH
	})
}

func TestNewDNSResolver(t *testing.T) {
	t.Run("default timeout when zero", func(t *testing.T) {
		resolver := NewDNSResolver("8.8.8.8", 0)
		assert.Equal(t, 10*time.Second, resolver.timeout)
	})

	t.Run("custom timeout", func(t *testing.T) {
		resolver := NewDNSResolver("8.8.8.8", 5*time.Second)
		assert.Equal(t, 5*time.Second, resolver.timeout)
	})

	t.Run("appends port 53 when missing", func(t *testing.T) {
		resolver := NewDNSResolver("1.1.1.1", 0)
		assert.Equal(t, "1.1.1.1:53", resolver.nameserver)
	})

	t.Run("does not append port when already present", func(t *testing.T) {
		resolver := NewDNSResolver("1.1.1.1:5353", 0)
		assert.Equal(t, "1.1.1.1:5353", resolver.nameserver)
	})

	t.Run("system keyword uses system resolver", func(t *testing.T) {
		resolver := NewDNSResolver("system", 0)
		// It should resolve to either a system DNS or fallback 8.8.8.8:53.
		assert.Contains(t, resolver.nameserver, ":53")
	})

	t.Run("empty nameserver uses system resolver", func(t *testing.T) {
		resolver := NewDNSResolver("", 0)
		assert.Contains(t, resolver.nameserver, ":53")
	})
}
