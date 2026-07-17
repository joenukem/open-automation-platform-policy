# Deny promotion of unsigned content into the trusted repository.
package platform

import rego.v1

deny contains d if {
	input.action == "content.promote"
	not input.content.signed
	d := {
		"policy": "unsigned-content",
		"reason": sprintf("collection %v is not signed and cannot be promoted to the trusted repository", [input.content.name]),
	}
}
