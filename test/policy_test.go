package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xbsoftware/wfs-local"
)

func TestReadPolicy(t *testing.T) {
	policy := wfs.ReadOnlyPolicy{}

	assert.True(t, policy.Comply("/any", wfs.ReadOperation))
	assert.False(t, policy.Comply("/any", wfs.WriteOperation))
}

func TestDenyPolicy(t *testing.T) {
	policy := wfs.DenyPolicy{}
	assert.False(t, policy.Comply("/any", wfs.ReadOperation))
	assert.False(t, policy.Comply("/any", wfs.WriteOperation))
}

func TestAllowPolicy(t *testing.T) {
	policy := wfs.AllowPolicy{}
	assert.True(t, policy.Comply("/any", wfs.ReadOperation))
	assert.True(t, policy.Comply("/any", wfs.WriteOperation))
}

func TestForceRootPolicy(t *testing.T) {
	policy := wfs.ForceRootPolicy{Root: "/sandbox"}
	assert.False(t, policy.Comply("/any", wfs.ReadOperation))
	assert.True(t, policy.Comply("/sandbox/", wfs.ReadOperation))
	assert.True(t, policy.Comply("/sandbox/any", wfs.ReadOperation))

	assert.False(t, policy.Comply("/any", wfs.WriteOperation))
	assert.True(t, policy.Comply("/sandbox/", wfs.WriteOperation))
	assert.True(t, policy.Comply("/sandbox/any", wfs.WriteOperation))
}

func TestCombinedPolicy(t *testing.T) {
	policy := wfs.CombinedPolicy{}
	assert.True(t, policy.Comply("/any", wfs.ReadOperation))

	policy = wfs.CombinedPolicy{Policies: []wfs.Policy{
		wfs.AllowPolicy{},
	}}
	assert.True(t, policy.Comply("/any", wfs.ReadOperation))

	policy = wfs.CombinedPolicy{Policies: []wfs.Policy{
		wfs.AllowPolicy{},
		wfs.ReadOnlyPolicy{},
	}}
	assert.True(t, policy.Comply("/any", wfs.ReadOperation))
	assert.False(t, policy.Comply("/any", wfs.WriteOperation))

	policy = wfs.CombinedPolicy{Policies: []wfs.Policy{
		wfs.ReadOnlyPolicy{},
		wfs.ForceRootPolicy{Root: "/sandbox"},
	}}
	assert.False(t, policy.Comply("/any", wfs.ReadOperation))
	assert.True(t, policy.Comply("/sandbox/", wfs.ReadOperation))
	assert.False(t, policy.Comply("/sandbox/any", wfs.WriteOperation))
}
