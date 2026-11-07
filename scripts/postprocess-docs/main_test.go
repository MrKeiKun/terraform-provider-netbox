package main

import "testing"

func TestProcess(t *testing.T) {
	const page = `---
page_title: "netbox_tag Resource - terraform-provider-netbox"
subcategory: ""
description: |-
  :meta:subcategory:Extras:From the official documentation.
---

# netbox_tag (Resource)

:meta:subcategory:Extras:From the official documentation.
`
	const want = `---
page_title: "netbox_tag Resource - terraform-provider-netbox"
subcategory: "Extras"
description: |-
  From the official documentation.
---

# netbox_tag (Resource)

From the official documentation.
`

	got, changed := process(page)
	if !changed || got != want {
		t.Fatalf("changed=%v, got:\n%s", changed, got)
	}

	again, changed := process(got)
	if changed || again != got {
		t.Fatalf("second run must not change the page, changed=%v", changed)
	}
}

func TestProcessKeepsTemplateSubcategory(t *testing.T) {
	const page = `---
subcategory: "Data Center Inventory Management (DCIM)"
description: |-
  :meta:subcategory:Other:Text
---
:meta:subcategory:Other:Text
`
	const want = `---
subcategory: "Data Center Inventory Management (DCIM)"
description: |-
  Text
---
Text
`
	if got, _ := process(page); got != want {
		t.Fatalf("got:\n%s", got)
	}
}

func TestProcessWithoutMarker(t *testing.T) {
	const page = "---\nsubcategory: \"\"\n---\nText\n"
	if got, changed := process(page); changed || got != page {
		t.Fatalf("page without marker must stay as is, changed=%v", changed)
	}
}
