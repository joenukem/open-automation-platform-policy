# Aggregation entrypoint for the platform policy service.
#
# Every policy pack is in package `platform` and adds objects to the shared
# `deny` set. The action is allowed only when no pack denies it. The gateway
# pre-launch hook and the MCP server both query `data.platform.decision`.
package platform

import rego.v1

default allow := false

allow if count(deny) == 0

decision := {
	"allow": allow,
	"denials": [d | some d in deny],
}
