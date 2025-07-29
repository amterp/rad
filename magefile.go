//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const BIN_DIR = "bin"

var (
	Default      = Auto
	goexe        = "go"
	gofmtexe     = "gofmt"
	goimportsexe = "goimports"
)

func init() {
	if exe := os.Getenv("GOEXE"); exe != "" {
		goexe = exe
		gofmtexe = "gofmt.exe"
		goimportsexe = "goimports.exe"
	}
}

func Auto() error {
	for _, step := range []func() error{Generate, Format, Build, Install} {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func All() error {
	for _, step := range []func() error{Generate, Format, Build, Test} {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func Generate() error {
	fmt.Println("⚙️ Running generators...")
	if err := runCmd(goexe, "run", filepath.Join(".", "function-metadata", "extract.go")); err != nil {
		return err
	}
	return runCmd("mv", filepath.Join(".", "functions.txt"), filepath.Join(".", "lsp-server", "com", "embedded"))
}

func Format() error {
	fmt.Println("⚙️ Formatting files...")
	if err := runCmd(gofmtexe, "-s", "-w", "."); err != nil {
		return err
	}
	return runCmd(goimportsexe, "-w", ".")
}

func Build() error {
	fmt.Println("⚙️ Building...")
	if err := runCmd("mkdir", "-p", filepath.Join(".", BIN_DIR)); err != nil {
		return err
	}
	return runCmd(goexe, "build", "-o", filepath.Join(".", BIN_DIR, "radd"))
}

func Test() error {
	fmt.Println("⚙️ Running tests...")
	return runCmd(goexe, "test", filepath.FromSlash("./core/testing"))
}

func Install() error {
	fmt.Println("⚙️ Installing...")
	radDevBinaryLocation := filepath.Join(".", BIN_DIR, "radd")
	locationToInstallRadBinary := filepath.Join(os.Getenv("GOBIN"), "rad")
	return runCmd("cp", radDevBinaryLocation, locationToInstallRadBinary)
}

func Clean() error {
	fmt.Println("⚙️ Cleaning...")
	return os.RemoveAll(filepath.Join(".", BIN_DIR))
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
