package nifi

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestClientUserCreate(t *testing.T) {
	config := Config{
		Host:          "127.0.0.1:9443",
		ApiPath:       "nifi-api",
		AdminCertPath: "/opt/nifi-toolkit/target/nifi-admin.pem",
		AdminKeyPath:  "/opt/nifi-toolkit/target/nifi-admin.key",
	}
	client := NewClient(config)

	user := User{
		Revision: Revision{
			Version: 0,
		},
		Component: UserComponent{
			ParentGroupId: "root",
			Identity:      "test_user",
			Position: &Position{
				X: 0,
				Y: 0,
			},
		},
	}
	err := client.CreateUser(&user)
	assert.Equal(t, err, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(user.Component.Id)
	}
	assert.NotEmpty(t, user.Component.Id)
	user2, err2 := client.GetUser(user.Component.Id)
	assert.Equal(t, err2, nil)
	log.Println(user2.Component.Id)
	assert.NotEmpty(t, user2.Component.Id)

	err = client.DeleteUser(user2)
	assert.Equal(t, err, nil)
}

func TestClientUserSearch(t *testing.T) {
	config := Config{
		Host:          "127.0.0.1:9443",
		ApiPath:       "nifi-api",
		AdminCertPath: "/opt/nifi-toolkit/target/nifi-admin.pem",
		AdminKeyPath:  "/opt/nifi-toolkit/target/nifi-admin.key",
	}
	client := NewClient(config)

	userIds, err := client.GetUserIdsWithIdentity("test_user")
	log.Println(fmt.Sprintf("%s,%v", userIds, err))
}

func TestClientGroupCreate(t *testing.T) {
	config := Config{
		Host:          "127.0.0.1:9443",
		ApiPath:       "nifi-api",
		AdminCertPath: "/opt/nifi-toolkit/target/nifi-admin.pem",
		AdminKeyPath:  "/opt/nifi-toolkit/target/nifi-admin.key",
	}
	client := NewClient(config)
	user1 := User{
		Revision: Revision{
			Version: 0,
		},
		Component: UserComponent{
			ParentGroupId: "",
			Identity:      "test_grp_usr9",
			Position: &Position{
				X: 0,
				Y: 0,
			},
		},
	}
	err := client.CreateUser(&user1)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(user1.Component.Id)
	}
	assert.NotEmpty(t, user1.Component.Id)
	users := []Tenant{}
	users = append(users, *user1.ToTenant())

	group := Group{
		Revision: Revision{
			Version: 0,
		},
		Component: GroupComponent{
			ParentGroupId: "root",
			Identity:      "test_group2",
			Position: &Position{
				X: 0,
				Y: 0,
			},
			Users: users,
		},
	}
	b, _ := json.Marshal(group)
	// Convert bytes to string.
	s := string(b)
	log.Println(s)
	err = client.CreateGroup(&group)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(group.Component.Id)
	}
	assert.NotEmpty(t, group.Component.Id)
	group2, err2 := client.GetGroup(group.Component.Id)
	assert.Equal(t, err2, nil)
	log.Println(group2.Component.Id)
	assert.NotEmpty(t, group2.Component.Id)

	err = client.DeleteGroup(group2)
	assert.Equal(t, err, nil)
	err = client.DeleteUser(&user1)
	assert.Equal(t, err, nil)

}

func TestUserToSchema(t *testing.T) {
	r := ResourceUser()
	timeouts := &schema.ResourceTimeout{
		Create: schema.DefaultTimeout(40 * time.Minute),
		Update: schema.DefaultTimeout(80 * time.Minute),
		Delete: schema.DefaultTimeout(40 * time.Minute),
	}

	r.Timeouts = timeouts
	d := &schema.ResourceData{}
	d.SetId("abcdefg")

	//client := NewClient(config)
	user1 := User{
		Revision: Revision{
			Version: 0,
		},
		Component: UserComponent{
			ParentGroupId: "dummy",
			Identity:      "test_grp_usr9",
			Position: &Position{
				X: 0,
				Y: 0,
			},
		},
	}
	UserToSchema(d, &user1)

	actual := d.State()
	log.Println(actual)

	log.Println(d.Get("component"))
}

func TestGroupToSchema(t *testing.T) {
	r := ResourceGroup()
	timeouts := &schema.ResourceTimeout{
		Create: schema.DefaultTimeout(40 * time.Minute),
		Update: schema.DefaultTimeout(80 * time.Minute),
		Delete: schema.DefaultTimeout(40 * time.Minute),
	}

	r.Timeouts = timeouts
	d := &schema.ResourceData{}
	d.SetId("abcdefg")

	//client := NewClient(config)
	group1 := Group{
		Revision: Revision{
			Version: 0,
		},
		Component: GroupComponent{
			ParentGroupId: "dummy",
			Identity:      "test_grp",
			Position: &Position{
				X: 0,
				Y: 0,
			},
		},
	}
	GroupToSchema(d, &group1)

	actual := d.State()
	log.Println(actual)

	log.Println(d.Get("component"))
}

func TestClientRemoteProcessGroupCreate(t *testing.T) {
	config := Config{
		Host:          "127.0.0.1:9443",
		ApiPath:       "nifi-api",
		AdminCertPath: "/opt/nifi-toolkit/target/nifi-admin.pem",
		AdminKeyPath:  "/opt/nifi-toolkit/target/nifi-admin.key",
	}
	client := NewClient(config)

	processGroup := RemoteProcessGroup{
		Revision: Revision{
			Version: 0,
		},
		Component: RemoteProcessGroupComponent{
			ParentGroupId: "root",
			Name:          "test_remote_pg",
			Position: Position{
				X: 0,
				Y: 0,
			},
			TargetUris:        "https://localhost:9443/nifi",
			TransportProtocol: "http",
		},
	}
	client.CreateRemoteProcessGroup(&processGroup)
	assert.NotEmpty(t, processGroup.Component.Id)

	processGroup2, err := client.GetRemoteProcessGroup(processGroup.Component.Id)
	assert.Equal(t, err, nil)
	assert.NotEmpty(t, processGroup2.Component.Id)

	processGroup.Component.Name = "test_remote_pg2"
	err = client.UpdateRemoteProcessGroup(&processGroup)
	assert.Equal(t, err, nil)
}

func TestClientInputPortCreate(t *testing.T) {
	config := Config{
		Host:          "127.0.0.1:9443",
		ApiPath:       "nifi-api",
		AdminCertPath: "/opt/nifi-toolkit/target/nifi-admin.pem",
		AdminKeyPath:  "/opt/nifi-toolkit/target/nifi-admin.key",
	}
	client := NewClient(config)

	inputPort := Port{
		Revision: Revision{
			Version: 0,
		},
		Component: PortComponent{
			ParentGroupId: "root",
			Name:          "test_input_port",
			Position: Position{
				X: 0,
				Y: 0,
			},
			PortType: "INPUT_PORT",
		},
	}
	client.CreatePort(&inputPort)
	assert.NotEmpty(t, inputPort.Component.Id)

	inputPort2, err := client.GetPort(inputPort.Component.Id, inputPort.Component.PortType)
	assert.Equal(t, err, nil)
	assert.NotEmpty(t, inputPort2.Component.Id)

	inputPort.Component.Name = "test_input_port2"
	err = client.UpdatePort(&inputPort)
	assert.Equal(t, err, nil)
}
