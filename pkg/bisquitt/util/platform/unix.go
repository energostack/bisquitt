//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris zos

package platform

import (
	"os/user"
	"strconv"
	"syscall"
)

func HasSetUser() bool {
	return true
}

func HasSetGroup() bool {
	return true
}

func lookupUser(username string) (int, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return 0, err
	}
	uid, err := strconv.ParseInt(u.Uid, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(uid), nil
}

func lookupGroup(name string) (int, error) {
	g, err := user.LookupGroup(name)
	if err != nil {
		return 0, err
	}
	gid, err := strconv.ParseInt(g.Gid, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(gid), nil
}

func SetUser(username string) error {
	uid, err := lookupUser(username)
	if err != nil {
		return err
	}
	return syscall.Setuid(uid)
}

func SetGroup(name string) error {
	gid, err := lookupGroup(name)
	if err != nil {
		return err
	}
	return syscall.Setgid(gid)
}

func GetCurrentUser() (*user.User, error) {
	uid := syscall.Getuid()
	return user.LookupId(strconv.FormatInt(int64(uid), 10))
}

func GetCurrentGroup() (*user.Group, error) {
	gid := syscall.Getgid()
	return user.LookupGroupId(strconv.FormatInt(int64(gid), 10))
}
