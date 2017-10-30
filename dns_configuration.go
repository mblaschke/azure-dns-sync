package main

import (
	"log"
	"fmt"
	"net"
	"errors"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/bogdanovich/dns_resolver"
	"github.com/Azure/azure-sdk-for-go/arm/dns"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	// default resolver using public dns servers from google
	defaultResolver = dns_resolver.New([]string{"8.8.8.8", "8.8.4.4"})
)

type DnsConfigurationItem struct {
	Name string `yaml:"name"`
	DnsServer []string `yaml:"dns"`
	Azure struct {
		Zone string `yaml:"zone"`
		Name string `yaml:"name"`
		Ttl int64 `yaml:"ttl"`
		ResourceGroup string `yaml:"resourceGroup"`
	} `yaml:"azure"`

	resolver *dns_resolver.DnsResolver `yaml:"-"`
}

type DnsConfiguration struct {
	Entries []DnsConfigurationItem

	Default struct {
		ResourceGroup string `yaml:"resourceGroup"`
		Zone string `yaml:"zone"`
		Ttl int64 `yaml:"ttl"`
	}

	azureClient dns.RecordSetsClient `yaml:"-"`
}

// create new dns configuration object
// and parse yaml file
func (conf *DnsConfiguration) ReadFromConfig(path string) (error error) {
	// read yaml file
	ymlData, err := ioutil.ReadFile(path)
	if err != nil {
		error = err
		return
	}

	// parse yaml
	if err := yaml.Unmarshal([]byte(ymlData), &conf); err != nil {
		error = err
		return
	}

	// process entries
	for key, entry := range conf.Entries {
		// default resource group
		if entry.Azure.ResourceGroup == "" {
			entry.Azure.ResourceGroup = conf.Default.ResourceGroup
		}

		// default zone
		if entry.Azure.Zone == "" {
			entry.Azure.Zone = conf.Default.Zone
		}

		// default ttl
		if entry.Azure.Ttl == 0 {
			entry.Azure.Ttl = conf.Default.Ttl
		}

		if conf.Entries[key].Name == "" {
			error = errors.New("Name cannot be empty")
			return
		}

		if entry.Azure.Name == "" {
			error = errors.New("azure.name cannot be empty")
			return
		}

		if entry.Azure.Zone == "" {
			error = errors.New("azure.zone cannot be empty")
			return
		}

		if entry.Azure.Ttl == 0 {
			error = errors.New("azure.ttl cannot be empty")
			return
		}

		if entry.Azure.ResourceGroup == "" {
			error = errors.New("azure.resourcegroup cannot be empty")
			return
		}

		// set resolver
		if len(entry.DnsServer) >= 1 {
			entry.resolver = dns_resolver.New(entry.DnsServer)
		} else {
			entry.resolver = defaultResolver
		}

		// update entry
		conf.Entries[key] = entry
	}

	return
}

// set azure dns recordset client
func (conf *DnsConfiguration) SetAzureClient(client dns.RecordSetsClient) {
	conf.azureClient = client
}

// execute update run loop
func (conf *DnsConfiguration) Run() (error error) {
	for _, entry := range conf.Entries {
		log.Println(fmt.Sprintf("Processing %s (%s in zone %s)", entry.Name, entry.Azure.Name, entry.Azure.Zone))

		if err := entry.Execute(conf.azureClient); err != nil {
			error = err
			return
		}
	}

	return
}

// execute update run
func (entry *DnsConfigurationItem) Execute(azureClient dns.RecordSetsClient) (error error) {
	recordSet, err := entry.buildAzureDnsEntry()
	if err != nil {
		error = err
		return
	}

	log.Println(fmt.Sprintf("   updating Azure DNS record %s in zone %s (RG:%s)", entry.Azure.Name, entry.Azure.Zone, entry.Azure.ResourceGroup))
	_, err = azureClient.CreateOrUpdate(entry.Azure.ResourceGroup, entry.Azure.Zone, entry.Azure.Name, dns.A, recordSet, "", "")
	if err != nil {
		log.Fatalf("Error creating DNS record: %s, %v", entry.Azure.Zone, err)
		return
	}
	return
}

// create azure dns recordset
func (entry *DnsConfigurationItem) buildAzureDnsEntry() (recordSet dns.RecordSet, error error) {
	var dnsRecords []dns.ARecord

	addressList, err := entry.lookup()
	if err != nil {
		error = err
		return
	}

	// translate address list
	for _, address := range addressList {
		record := dns.ARecord{
			Ipv4Address: to.StringPtr(address.String()),
		}
		dnsRecords = append(dnsRecords, record)
	}

	recordSet = dns.RecordSet{
		RecordSetProperties: &dns.RecordSetProperties{
			TTL: to.Int64Ptr(entry.Azure.Ttl),
			ARecords: &dnsRecords,
		},
	}

	return
}

// lookup hostname using resolver
func (entry *DnsConfigurationItem) lookup() (addressList []net.IP, error error) {
	log.Println(fmt.Sprintf("   resolving %s using %v", entry.Name, entry.resolver.Servers))
	addressList, err := entry.resolver.LookupHost(entry.Name)
	if err != nil {
		error = err
		return
	}

	log.Println(fmt.Sprintf("   resolved %s to %v", entry.Name, addressList))
	return
}
