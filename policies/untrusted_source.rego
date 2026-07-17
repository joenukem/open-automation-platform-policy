# Deny importing content from a source that is not on the trusted list.
package platform

import rego.v1

deny contains d if {
	input.action == "content.import"
	not trusted_source
	d := {
		"policy": "untrusted-source",
		"reason": sprintf("content source %v is not in the trusted source list %v", [input.content.source, data.config.trusted_sources]),
	}
}

trusted_source if {
	some src in data.config.trusted_sources
	contains(input.content.source, src)
}
