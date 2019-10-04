package okta

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/hashicorp/terraform/helper/acctest"
)

const baseTestProp = "firstName"

func sweepUserBaseSchema(client *testClient) error {
	_, _, err := client.artClient.Schemas.GetUserSchema()
	if err != nil {
		return err
	}
	var errorList []error

	return condenseError(errorList)
}

func TestAccOktaUserBaseSchema_crud(t *testing.T) {
	ri := acctest.RandInt()
	mgr := newFixtureManager(userBaseSchema)
	config := mgr.GetFixtures("basic.tf", ri, t)
	updated := mgr.GetFixtures("updated.tf", ri, t)
	usernamePattern := mgr.GetFixtures("username.tf", ri, t)
	resourceName := fmt.Sprintf("%s.%s", userBaseSchema, baseTestProp)
	loginResourceName := fmt.Sprintf("%s.login", userBaseSchema)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil, // can't delete base properties
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testOktaUserBaseSchemasExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "index", baseTestProp),
					resource.TestCheckResourceAttr(resourceName, "title", "First name"),
					resource.TestCheckResourceAttr(resourceName, "type", "string"),
					resource.TestCheckResourceAttr(resourceName, "permissions", "READ_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "min_length", "1"),
					resource.TestCheckResourceAttr(resourceName, "max_length", "50"),
				),
			},
			{
				Config: updated,
				Check: resource.ComposeTestCheckFunc(
					testOktaUserBaseSchemasExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "index", baseTestProp),
					resource.TestCheckResourceAttr(resourceName, "title", "First name"),
					resource.TestCheckResourceAttr(resourceName, "type", "string"),
					resource.TestCheckResourceAttr(resourceName, "required", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions", "READ_WRITE"),
					resource.TestCheckResourceAttr(resourceName, "min_length", "1"),
					resource.TestCheckResourceAttr(resourceName, "max_length", "50"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return errors.New("Failed to import schema into state")
					}

					return nil
				},
			},
			{
				Config: usernamePattern,
				Check: resource.ComposeTestCheckFunc(
					testOktaUserBaseSchemasExists(loginResourceName),
					resource.TestCheckResourceAttr(loginResourceName, "index", "login"),
					resource.TestCheckResourceAttr(loginResourceName, "title", "Username"),
					resource.TestCheckResourceAttr(loginResourceName, "type", "string"),
					resource.TestCheckResourceAttr(loginResourceName, "required", "true"),
					resource.TestCheckResourceAttr(loginResourceName, "permissions", "READ_ONLY"),
					resource.TestCheckResourceAttr(loginResourceName, "min_length", "5"),
					resource.TestCheckResourceAttr(loginResourceName, "max_length", "70"),
				),
			},
		},
	})
}

func testOktaUserBaseSchemasExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if exists, _ := testUserBaseSchemaExists(rs.Primary.ID); !exists {
			return fmt.Errorf("Failed to find %s", rs.Primary.ID)
		}
		return nil
	}
}

func testUserBaseSchemaExists(index string) (bool, error) {
	client := getClientFromMetadata(testAccProvider.Meta())
	subschema, _, err := client.Schemas.GetUserSubSchemaIndex(baseSchema)
	if err != nil {
		return false, fmt.Errorf("Error Listing User Subschema in Okta: %v", err)
	}
	for _, key := range subschema {
		if key == index {
			return true, nil
		}
	}

	return false, nil
}
