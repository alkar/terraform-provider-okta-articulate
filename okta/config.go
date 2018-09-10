package okta

import (
	"fmt"

	articulateOkta "github.com/articulate/oktasdk-go/okta"
)

// Config is a struct containing our provider schema values
// plus the okta client object
type Config struct {
	orgName  string
	domain   string
	apiToken string

	articulateOktaClient *articulateOkta.Client
}

func (c *Config) loadAndValidateArticulateSDK() error {

	client, err := articulateOkta.NewClientWithDomain(nil, c.orgName, c.domain, c.apiToken)
	if err != nil {
		return fmt.Errorf("[ERROR] Error creating Okta client: %v", err)
	}

	// quick test of our credentials by listing our default user profile schema
	url := fmt.Sprintf("meta/schemas/user/default")
	req, err := client.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("[ERROR] Error initializing test connect to Okta: %v", err)
	}
	_, err = client.Do(req, nil)
	if err != nil {
		return fmt.Errorf("[ERROR] Error testing connection to Okta. Please verify your credentials: %v", err)
	}

	// add our client object to Config
	c.articulateOktaClient = client
	return nil
}
