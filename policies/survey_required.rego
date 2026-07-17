# Require an enabled survey when launching a privileged job template.
package platform

import rego.v1

deny contains d if {
	input.action == "job.launch"
	input.job.privileged == true
	not input.job.survey_enabled
	d := {
		"policy": "survey-required",
		"reason": sprintf("privileged template %v requires an enabled survey", [input.job.template]),
	}
}
