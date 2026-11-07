package netbox

import (
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/stretchr/testify/assert"
)

func TestTagNames(t *testing.T) {
	tags := []netboxapi.TagRef{
		{Name: "Foo", Slug: "foo"},
		{Name: "Bar", Slug: "bar"},
	}

	assert.Equal(t, []string{"Foo", "Bar"}, tagNames(tags))
	assert.Equal(t, []string{}, tagNames(nil))
}
