# Restrict launches against production inventories to approved teams.
package platform

import rego.v1

prod_actions := {"job.launch", "workflow.launch"}

deny contains d if {
	input.action in prod_actions
	input.target.inventory in data.config.prod_inventories
	not actor_approved_for_prod
	d := {
		"policy": "prod-inventory",
		"reason": sprintf("inventory %v is production; actor teams %v are not in the approved set %v", [input.target.inventory, input.actor.teams, data.config.approved_prod_teams]),
	}
}

actor_approved_for_prod if {
	some t in input.actor.teams
	t in data.config.approved_prod_teams
}
