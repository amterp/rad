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
	Default      = All
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

func All() error {
	Generate()
	Format()
	Build()
	Test()
	return Install()
}

func Generate() error {
	fmt.Println("⚙️  Running generators...")
	cmd := exec.Command(goexe, "run", filepath.Join(".", "function-metadata", "extract.go"))
	mov := exec.Command("mv", filepath.Join(".", "functions.txt"), filepath.Join(".", "lsp-server", "com", "embedded"))
	cmd.Stdout = os.Stdout
	mov.Stdout = os.Stdout
	cmd.Run()
	return mov.Run()
}

func Format() error {
	fmt.Println("⚙️  Formatting files...")
	cmd := exec.Command(gofmtexe, "-s", "-w", ".")
	imp := exec.Command(goimportsexe, "-w", ",")
	cmd.Stdout = os.Stdout
	imp.Stdout = os.Stdout
	cmd.Run()
	return imp.Run()
}

func Build() error {
	fmt.Println("Building...")
	mdir := exec.Command("mkdir", "-p", filepath.Join(".", BIN_DIR))
	cmd := exec.Command(goexe, "build", "-o", filepath.Join(".", BIN_DIR, "radd"))
	mdir.Stdout = os.Stdout
	cmd.Stdout = os.Stdout
	mdir.Run()
	return cmd.Run()
}

func Test() error {
	fmt.Println("⚙️  Running tests...")
	cmd := exec.Command(goexe, "test", filepath.FromSlash("./core/testing"))
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func Install() error {
	fmt.Println("⚙️  Installing...")
	cmd := exec.Command("cp", filepath.Join(".", BIN_DIR, "radd"), filepath.Join(os.Getenv("GOROOT"), "bin", "rad"))
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll(filepath.Join(".", BIN_DIR))
}
