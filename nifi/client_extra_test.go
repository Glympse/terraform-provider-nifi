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
	err_create := client.CreateUser(&user)
	if err_create != nil {
		log.Fatal(err_create)
	} else {
		log.Println(user.Component.Id)
	}
	assert.NotEmpty(t, user.Component.Id)
	user2, err := client.GetUser(user.Component.Id)
	assert.Equal(t, err, nil)
	log.Println(user2.Component.Id)
	assert.NotEmpty(t, user2.Component.Id)
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
	err_create_user := client.CreateUser(&user1)
	if err_create_user != nil {
		log.Fatal(err_create_user)
	} else {
		log.Println(user1.Component.Id)
	}
	assert.NotEmpty(t, user1.Component.Id)

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
	err_create := client.CreateGroup(&group)
	if err_create != nil {
		log.Fatal(err_create)
	} else {
		log.Println(group.Component.Id)
	}
	assert.NotEmpty(t, group.Component.Id)
	group2, err := client.GetGroup(group.Component.Id)
	assert.Equal(t, err, nil)
	log.Println(group2.Component.Id)
	assert.NotEmpty(t, group2.Component.Id)

	// group2.Component.Users = users
	// b, _ = json.Marshal(group2)
	// // Convert bytes to string.
	// s = string(b)
	// log.Println(s)
	// err_update := client.UpdateGroup(group2)
	//
	// if err_update != nil {
	// 	log.Fatal(err_update)
	// } else {
	// 	log.Println("Update group success")
	// }
}
