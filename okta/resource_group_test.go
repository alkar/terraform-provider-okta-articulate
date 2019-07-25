package okta

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/okta/okta-sdk-golang/okta/query"
)

func sweepGroups(client *testClient) error {
	var errorList []error
	// Should never need to deal with pagination, limit is 10,000 by default
	groups, _, err := client.oktaClient.Group.ListGroups(&query.Params{Q: testResourcePrefix})
	if err != nil {
		return err
	}

	for _, s := range groups {
		if _, err := client.oktaClient.Group.DeleteGroup(s.Id); err != nil {
			errorList = append(errorList, err)
		}

	}
	return condenseError(errorList)
}

// https://github.com/articulate/terraform-provider-okta/issues/220
func TestAccOktaGroupImport(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := fmt.Sprintf("%s.test", oktaGroup)
	mgr := newFixtureManager(oktaGroup)
	config := mgr.GetFixtures("okta_group_with_users.tf", ri, t)

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return errors.New("Failed to import group into state")
					}

					return nil
				},
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("testAcc_%d", ri)),
					resource.TestCheckResourceAttr(resourceName, "users.#", "4"),
				),
			},
		},
	})
}

func TestAccOktaGroupCreate(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := fmt.Sprintf("%s.test", oktaGroup)
	mgr := newFixtureManager("okta_group")
	config := mgr.GetFixtures("okta_group.tf", ri, t)
	updatedConfig := mgr.GetFixtures("okta_group_updated.tf", ri, t)
	addUsersConfig := mgr.GetFixtures("okta_group_with_users.tf", ri, t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "testAcc")),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "testAccDifferent")),
			},
			{
				Config: addUsersConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("testAcc_%d", ri)),
					resource.TestCheckResourceAttr(resourceName, "users.#", "4"),
				),
			},
		},
	})
}
