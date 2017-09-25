package nifi

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestClientUserCreate(t *testing.T) {
	config := Config{
		Host:    "127.0.0.1:9443",
		ApiPath: "nifi-api",
		AdminCertPath: "/opt/nifi-toolkit/target/nifi-cert.pem",
		AdminKeyPath: "/opt/nifi-toolkit/target/nifi-key.key"
	}
	client := NewClient(config)

	user := User{
		Revision: Revision{
			Version: 0,
		},
		Component: UserComponent{
			ParentGroupId: "root",
			Identity:      "test_user",
			Position: Position{
				X: 0,
				Y: 0,
			},
		},
	}
	client.CreateUser(&user)
	assert.NotEmpty(t, user.Component.Id)

	processGroup2, err := client.GetUser(user.Component.Id)
	assert.Equal(t, err, nil)
	assert.NotEmpty(t, user.Component.Id)

}
