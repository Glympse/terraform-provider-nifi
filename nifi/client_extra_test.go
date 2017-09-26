package nifi

import (
	"encoding/json"
	"log"
	"testing"

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

	user2 := User{
		Revision: Revision{
			Version: 0,
		},
		Component: UserComponent{
			ParentGroupId: "",
			Identity:      "test_grp_usr10",
			Position: &Position{
				X: 0,
				Y: 0,
			},
		},
	}
	err = client.CreateUser(&user2)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(user2.Component.Id)
	}
	assert.NotEmpty(t, user2.Component.Id)

	users := []*Tenant{}
	users = append(users, user1.ToTenant())

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

	users = append(users, user2.ToTenant())

	group2.Component.Users = users
	b, _ = json.Marshal(group2)
	// Convert bytes to string.
	s = string(b)
	log.Println(s)
	err = client.UpdateGroup(group2)

	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Update group success")
	}

	err = client.DeleteGroup(group2)
	assert.Equal(t, err, nil)
	err = client.DeleteUser(&user1)
	assert.Equal(t, err, nil)
	err = client.DeleteUser(&user2)
	assert.Equal(t, err, nil)
}
