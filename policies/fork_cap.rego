# Cap the number of concurrent forks a job launch may request.
package platform

import rego.v1

deny contains d if {
	input.action == "job.launch"
	input.job.forks > data.config.max_forks
	d := {
		"policy": "fork-cap",
		"reason": sprintf("requested forks %v exceed the platform cap of %v", [input.job.forks, data.config.max_forks]),
	}
}
