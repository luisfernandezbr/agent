package eventapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPinpointWebhook(t *testing.T) {
	assert := assert.New(t)
	m := eventAPIManager{
		channel: "edge",
	}
	assert.False(m.IsPinpointWebhook(""))
	assert.True(m.IsPinpointWebhook("https://event.api.edge.pinpoint.com/hook/wpuuNktNNhQtjJKC2OyX%2F3WkPEng5MFQjMdjweWpLGqQ%2F5ROLcthHaPCgCRQG5ybaUhrLncO1tmjmsYYFbUbiW4R63LxUCNBPHAiS2wgXDK0PVg2y667AEb1FB6bFQVN9ZSrR8cZ8J86dcxwvNlPZYsDdPQDBGcvIAYeE5j600IOlJRqVc%2FGpDBVL5d6lIPjEeX1lC6YGkM6NvkdYyXE8WrEhZvqfxOuEOasy3dYPwibDuLZsWiRxDdwWm5B%2F0YqMlzzaDRI4P8HPGseA7F5+7JnHDxoKEMZMcGW2X3I8qk=?integration_instance_id=c0b35e2adea4fd36"))
	assert.True(m.IsPinpointWebhook("https://webhook.api.edge.pinpoint.com/shared/repo/github/11111"))
	m.channel = "stable"
	assert.False(m.IsPinpointWebhook(""))
	assert.False(m.IsPinpointWebhook("https://event.api.edge.pinpoint.com/hook/wpuuNktNNhQtjJKC2OyX%2F3WkPEng5MFQjMdjweWpLGqQ%2F5ROLcthHaPCgCRQG5ybaUhrLncO1tmjmsYYFbUbiW4R63LxUCNBPHAiS2wgXDK0PVg2y667AEb1FB6bFQVN9ZSrR8cZ8J86dcxwvNlPZYsDdPQDBGcvIAYeE5j600IOlJRqVc%2FGpDBVL5d6lIPjEeX1lC6YGkM6NvkdYyXE8WrEhZvqfxOuEOasy3dYPwibDuLZsWiRxDdwWm5B%2F0YqMlzzaDRI4P8HPGseA7F5+7JnHDxoKEMZMcGW2X3I8qk=?integration_instance_id=c0b35e2adea4fd36"))
	assert.False(m.IsPinpointWebhook("https://webhook.api.edge.pinpoint.com/shared/repo/github/11111"))
	assert.True(m.IsPinpointWebhook("https://event.api.pinpoint.com/hook/VA2yT735Lbl9Vk8sg32AK655XF%2FFKW2lMH6dg1k6cs1YVF4TghpSHlmBuWq%2Fx00U4zib1X0d98WX8uJAtntDUOhZbJDCpS4T+vsJhgfWUqNIHGKXuQw9ioNSUrHNpmBXELHO3HMedWPbj%2FvsVnFver6bkIh+QhXkTADiinpcfnnRIyv7FS%2FAgG+PSYSyV%2F3MxP8iSWNmInRWCo%2FkjOFw9dTV3nYDT4CE1l+xJZ4oQXHVG+9tPtcYUA1obz5u7Bi%2FlyJVNg=="))
	m.channel = "dev"
	assert.False(m.IsPinpointWebhook(""))
	assert.True(m.IsPinpointWebhook("https://webhook.api.ppoint.io:8454/shared/repo/github/11111"))
}
