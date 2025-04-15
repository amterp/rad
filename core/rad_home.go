package core

const RAD_HOME_DIR = "RAD_HOME_DIR"

var RadHomeInst *RadHome

type RadHome struct {
	HomeDir string
}

func NewRadHome(home string) *RadHome {
	return &RadHome{
		HomeDir: home,
	}
}
