package borg

import (
	"fmt"
	"github.com/lxc/lxd/shared"
	"os/exec"
)

// each volume has it's own repo

/*

   let out = utils.tree()
   let borgCmd = []
   if (config.storage.sshpass) {
     borgCmd.push('sshpass', '-p', config.storage.sshpass)
   }

   borgCmd.push('borg')

   if (config.storage.repo) {
     out.evar('BORG_REPO', config.storage.repo)
   }
   if (config.storage.passphrase) {
     out.evar('BORG_PASSPHRASE', config.storage.passphrase)
   }
   if (config.storage.passcommand) {
     out.evar('BORG_PASSPHRASE', config.storage.passcommand)
   }
   out.evar('BORG_RSH', 'ssh -o StrictHostKeyChecking=no')

   let listCmd = borgCmd.slice(0)
   listCmd.push('list')
   let initCmd = borgCmd.slice(0)
   initCmd.push('init', '-e', 'none')
   out.if('! yes | ' + utils.shellEscape(listCmd) + ' >/dev/null 2>/dev/null', utils.shellEscape(initCmd))

   let createCmd = borgCmd.slice(0)
   createCmd.push('create', '--list', '--stats')
   if (config.create.exclude) {
     config.create.exclude.forEach(e => createCmd.push('--exclude', e))
   }
   if (config.create.excludedCaches) {
     createCmd.push('--exclude-caches')
   }
   if (config.extraArgs && config.extraArgs.create) {
     createCmd.push(...config.extraArgs.create)
   }
   createCmd.push('::' + config.create.name, '${LIST[@]}') // eslint-disable-line no-template-curly-in-string
   createCmd.unshift('safeexec')

   out // run create, if warning try again, otherwise continue. exit with error if final exit code non-zero
     .var('RUN_CREATE', 'true')
     .while('$RUN_CREATE', utils.tree()
       .append('yes | \\')
       .cmd(...createCmd)
       .if('[ $ex -ne 1 ]', 'RUN_CREATE=false'))
     .if('[ $ex -ne 0 ]', utils.tree().cmd('echo', 'Borg backup failed with $ex').cmd('exit', '$ex'))

   out.cmd('rm', '-rf', '${RM_LIST[@]}') // eslint-disable-line no-template-curly-in-string

   let pruneCmd = borgCmd.slice(0)
   pruneCmd.push('prune', '--list', '--stats')
   for (const opt in config.prune) { // eslint-disable-line guard-for-in
     pruneCmd.push('--' + PRUNE_OPT_TRANSLATE[opt], config.prune[opt])
   }
   if (config.extraArgs && config.extraArgs.prune) {
     pruneCmd.push(...config.extraArgs.prune)
   }

   out.cmd(...pruneCmd)

   return utils.wrap('backup', 'backup', {cron: out.str(), priority: 1000})

 */

// RUnBorg spawns a borgbackup instance with predefined environment and arguments
func RunBorg(repo map[string]string, extraBorgEnv map[string]string, borgArgs ...string) (string, error) {
	/* borgVerbosity := "-q"
	if Debug {
		borgVerbosity = "-vi"
	} */

	borgEnv := map[string]string{
		"BORG_REPO": repo["repo"],
		// BORG_PASSPHRASE:
		"BORG_RSH": "ssh -o StrictHostKeyChecking=no",
	}

	// add passphrase command to env if used
	if repo["passphrase"] != "" {
		borgEnv["BORG_PASSPHRASE"] = repo["passphrase"]
	}

	// add key as -i arg to ssh if used
	if repo["key"] != "" {
		borgEnv["BORG_RSH"] = borgEnv["BORG_RSH"] + " -i" + repo["key"]
	}

	// concat borgEnv
	for k, v := range extraBorgEnv {
		borgEnv[k] = v
	}

	args := []string{}

	// set environment
	for key, value := range borgEnv {
		args = append(args, fmt.Sprintf("%s=%s", key, value))
	}

	// set command
	args = append(args, "borg-wrapped")

	// append arguments
	args = append(args, borgArgs...)

	msg, err := shared.RunCommand("env")
	if err != nil {
		runError, ok := err.(shared.RunError)
		if ok {
			exitError, ok := runError.Err.(*exec.ExitError)
			if ok {
				if exitError.ExitCode() == 24 {
					return msg, nil
				}
			}
		}
		return msg, err
	}

	return msg, nil
}

// BorgCreate creates a borgbackup in the specified repo of the specified folder
func BorgCreate(repo map[string]string, name string, sourceFolder string) (string, error) {
	return RunBorg(
		repo,
		map[string]string{
			"SET_CWD": sourceFolder,
		},
		"create", "--list", "--stats", "::" + name,
		".")
}

func BorgInit(repo map[string]string) (string, error) {
	return RunBorg(repo, map[string]string{
		"YES_PIPE": "1",
	}, "init", "-e", "none")
}

// BorgPrepare checks if the repo exists and initializes it if it doesn't
func BorgPrepare(repo map[string]string) error {
	_, err := RunBorg(repo, map[string]string{
		"YES_PIPE": "1",
	}, "list")

	if err != nil {
		_, err = BorgInit(repo)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

// BorgRestore restores a given archive to a given folder
func BorgRestore(repo map[string]string, name string, destFolder string) (string, error) {
	return RunBorg(repo, map[string]string {
		"SET_CWD": destFolder,
	}, "extract", "::" + name, ".")
}

// BorgDelete removes a given archive
func BorgDelete(repo map[string]string, name string) (string, error) {
	return RunBorg(repo, map[string]string {
		"YeS_PIPE": "1",
	}, "delete", "::" + name, "--force")
}