package main

import (
	"github.com/Azure/azure-sdk-for-go/arm/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/go-autorest/autorest/azure"
	"io/ioutil"
	"encoding/json"
	"errors"
	"log"
	"fmt"
)

type AzureConfiguration struct {
	TenantId string `json:"tenantId,omitempty"`
	SubscriptionId string `json:"subscriptionId,omitempty"`
	AadClientId string `json:"aadClientId,omitempty"`
	AadClientSecret string `json:"aadClientSecret,omitempty"`
	DnsZoneResourceGroup string `json:"-"`
}

func (config *AzureConfiguration) ReadFromConfig(path string) (error error) {
	log.Println(fmt.Sprintf("Parsing AZURE configuration from %s", path))

	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		error = err
		return
	}

	err = json.Unmarshal([]byte(jsonData), &config)
	if err != nil {
		error = err
		return
	}

	if config.TenantId == "" {
		error = errors.New("tenantId is empty")
	}

	if config.SubscriptionId == "" {
		error = errors.New("subscriptionId is empty")
	}

	if config.AadClientId == "" {
		error = errors.New("aadClientId is empty")
	}

	if config.AadClientSecret == "" {
		error = errors.New("aadClientSecret is empty")
	}

	return
}

func (config *AzureConfiguration) GetClient() (rc dns.RecordSetsClient, error error) {
	log.Println("Fetching AZURE service principal token")

	c := map[string]string{
		"AZURE_CLIENT_ID":       config.AadClientId,
		"AZURE_CLIENT_SECRET":   config.AadClientSecret,
		"AZURE_SUBSCRIPTION_ID": config.SubscriptionId,
		"AZURE_TENANT_ID":       config.TenantId,
	}

	spt, err := helpers.NewServicePrincipalTokenFromCredentials(c, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		error = err
		return
	}

	rc = dns.NewRecordSetsClient(c["AZURE_SUBSCRIPTION_ID"])
	rc.Authorizer = autorest.NewBearerAuthorizer(spt)

	log.Println(" * successfull")

	return
}
