package nifi

import (
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
