package gitlib

import (
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func CloneRepoWithSSH(path, repoURL string, sshKey []byte) error {
	// Load private key
	publicKeys, err := ssh.NewPublicKeys(
		"git",  // SSH user
		sshKey, // github ssh public key
		"",     // passphrase (empty if none)
	)
	if err != nil {
		return err
	}

	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:  repoURL,
		Auth: publicKeys,
	})

	return err
}
