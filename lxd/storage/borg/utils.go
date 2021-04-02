package borg

import (
	"github.com/lxc/lxd/shared"
)

func IsBorg(config map[string]string) bool {
	return shared.IsTrue(config["borg.enabled"])
}

func GetBorgRepo(config map[string]string, volName string) map[string]string {
	return map[string]string { // TODO: review all is named right
		"repo": config["borg.repo"] + "/" + volName,
		"passphrase": config["borg.sshpass"],
		"key": config["borg.sshkey"],
	}
}