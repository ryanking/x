//+build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func g(cmd ...string) error {
	return sh.Run("go", cmd...)
}

func Build() error {
	mg.Deps(Mod)
	return g("build", ".")
}

func Test() error {
	mg.Deps(Mod)
	return g("test", "./...")
}

func Mod() error {
	return g("mod", "tidy")
}

func Install() error {
	return g("install", ".")
}

func Run() error {
	mg.Deps(Mod)
	return g("run", ".")
}
