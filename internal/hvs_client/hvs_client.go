package hvs_client

import (
	"log"
	"os"

	vs "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/config"
	"github.com/hashicorp/hcp-sdk-go/httpclient"
)

type HVSClient struct {
	OrgID     string
	ProjectID string
	Client    vs.ClientService
}

func NewHVSClient() *HVSClient {
	c := new(HVSClient)

	hcpConfig, err := config.NewHCPConfig(
		config.FromEnv(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Construct HTTP client config
	httpclientConfig := httpclient.Config{
		HCPConfig: hcpConfig,
	}

	// Initialize SDK http client
	cl, err := httpclient.New(httpclientConfig)
	if err != nil {
		log.Fatal(err)
	}

	// These IDs can be obtained from the portal URL
	c.OrgID = os.Getenv("HCP_ORGANIZATION_ID")
	c.ProjectID = os.Getenv("HCP_PROJECT_ID")

	c.Client = vs.New(cl, nil)

	return c
}

func (c *HVSClient) GetSecret(appName string, secretName string) (string, error) {
	reqParams := vs.NewOpenAppSecretParams()
	reqParams.LocationOrganizationID = c.OrgID
	reqParams.LocationProjectID = c.ProjectID
	reqParams.AppName = appName
	reqParams.SecretName = secretName

	hostResp, err := c.Client.OpenAppSecret(reqParams, nil)
	if err != nil {
		return "", err
	}

	return hostResp.Payload.Secret.Version.Value, nil
}
