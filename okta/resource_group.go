package okta

import (
	"net/http"

	"github.com/okta/okta-sdk-golang/okta"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/okta/okta-sdk-golang/okta/query"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Exists: resourceGroupExists,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Group names",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Group description",
			},
			"manage_users": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "By flipping this on it will manage users for a group based on users attribute",
			},
			"users": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Users associated with the group. This can also be done per user.",
			},
		},
	}
}

func buildGroup(d *schema.ResourceData) *okta.Group {
	return &okta.Group{
		Profile: &okta.GroupProfile{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		},
	}
}

func resourceGroupCreate(d *schema.ResourceData, m interface{}) error {
	group := buildGroup(d)
	responseGroup, _, err := getOktaClientFromMetadata(m).Group.CreateGroup(*group)
	if err != nil {
		return err
	}

	d.SetId(responseGroup.Id)
	if err := updateGroupUsers(d, m); err != nil {
		return err
	}

	return resourceGroupRead(d, m)
}

func resourceGroupExists(d *schema.ResourceData, m interface{}) (bool, error) {
	g, err := fetchGroup(d, m)

	return err == nil && g != nil, err
}

func resourceGroupRead(d *schema.ResourceData, m interface{}) error {
	g, err := fetchGroup(d, m)
	if err != nil {
		return err
	}

	d.Set("name", g.Profile.Name)
	d.Set("description", g.Profile.Description)
	if err := syncGroupUsers(d, m); err != nil {
		return err
	}

	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, m interface{}) error {
	group := buildGroup(d)
	_, _, err := getOktaClientFromMetadata(m).Group.UpdateGroup(d.Id(), *group)
	if err != nil {
		return err
	}

	if err := updateGroupUsers(d, m); err != nil {
		return err
	}

	return resourceGroupRead(d, m)
}

func resourceGroupDelete(d *schema.ResourceData, m interface{}) error {
	_, err := getOktaClientFromMetadata(m).Group.DeleteGroup(d.Id())

	return err
}

func fetchGroup(d *schema.ResourceData, m interface{}) (*okta.Group, error) {
	g, resp, err := getOktaClientFromMetadata(m).Group.GetGroup(d.Id(), &query.Params{})

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return g, err
}

func syncGroupUsers(d *schema.ResourceData, m interface{}) error {
	// Only sync when the manage_users property is true
	if !d.Get("manage_users").(bool) {
		return nil
	}
	userIdList, err := listGroupUserIds(m, d.Id())
	if err != nil {
		return err
	}

	return d.Set("users", convertStringSetToInterface(userIdList))
}

func updateGroupUsers(d *schema.ResourceData, m interface{}) error {
	// Only sync when the manage_users property is true
	if !d.Get("manage_users").(bool) {
		return nil
	}

	client := getOktaClientFromMetadata(m)
	existingUserList, _, err := client.Group.ListGroupUsers(d.Id(), nil)
	if err != nil {
		return err
	}

	rawArr := d.Get("users").(*schema.Set).List()
	userIdList := make([]string, len(rawArr))

	for i, ifaceId := range rawArr {
		userId := ifaceId.(string)
		userIdList[i] = userId

		if !containsUser(existingUserList, userId) {
			resp, err := client.Group.AddUserToGroup(d.Id(), userId)
			if err != nil {
				return responseErr(resp, err)
			}
		}
	}

	for _, user := range existingUserList {
		if !contains(userIdList, user.Id) {
			err := suppressErrorOn404(client.Group.RemoveGroupUser(d.Id(), user.Id))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func containsUser(users []*okta.User, id string) bool {
	for _, user := range users {
		if user.Id == id {
			return true
		}
	}
	return false
}
