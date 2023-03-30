package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEqual(t *testing.T) {
	theCore := EasyHttpSpec{
		Host:           "testhost",
		Replicas:       nil,
		Image:          "testimage",
		ImageTag:       "1.0",
		Port:           1234,
		Env:            map[string]string{"PORT": "1111", "PORT2": "1234"},
		CertManInssuer: "local.issuer",
		Path:           "/app",
	}
	cpy := theCore.DeepCopy()
	assert.True(t, theCore.IsEqual(cpy))

	var oneValue int32 = 1
	cpy.Replicas = &oneValue
	assert.True(t, theCore.IsEqual(cpy))

	cpy.Env["PORT2"] = "12345"
	assert.False(t, theCore.IsEqual(cpy))

	cpy = theCore.DeepCopy()
	cpy.Env["PORT3"] = "12345"
	assert.False(t, theCore.IsEqual(cpy))

	cpy = theCore.DeepCopy()
	cpy.Host = "12345"
	assert.False(t, theCore.IsEqual(cpy))

	cpy = theCore.DeepCopy()
	cpy.Image = "notthesamse"
	assert.False(t, theCore.IsEqual(cpy))

	cpy = theCore.DeepCopy()
	cpy.Env = make(map[string]string)
	theCore.Env = make(map[string]string)
	assert.True(t, theCore.IsEqual(cpy))

}
