// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	// run first time ignoring output, to make sure modules list is up
	// to date
	_, err := exec.Command("go", "list", "-f", "'{{.Standard}}#{{.ImportPath}}#{{.Module}}#{{.Imports}}'", "-deps").Output()
	if err != nil {
		log.Fatal(err)
	}

	// now, run again, this time getting the name of the main package
	out, err := exec.Command("go", "list").Output()
	if err != nil {
		log.Fatal(err)
	}
	mainPackage := strings.TrimSpace(string(out))

	// and now, run again, this time capturing output
	out, err = exec.Command("go", "list", "-f", "'{{.Standard}}#{{.ImportPath}}#{{.Module}}#{{.Imports}}'", "-deps").Output()
	if err != nil {
		log.Fatal(err)
	}

	// and parse that output into a series of package info records
	pis, err := parseGoListOutput(out, mainPackage)
	if err != nil {
		log.Fatal(err)
	}

	// and gather modules
	mods := getModules(pis)

	// and print those details for NON-standard pkgs only
	for _, pi := range pis {
		if !pi.isStandard {
			if pi.isMainPackage {
				fmt.Printf("*** ")
			}
			fmt.Printf("%s (%s %s): ", pi.importPath, pi.modName, pi.modVersion)
			anyDeps := false
			for _, dep := range pi.deps {
				depPi := pis[dep]
				if !depPi.isStandard {
					fmt.Printf("\n  - %s", depPi.importPath)
					anyDeps = true
				}
			}
			if !anyDeps {
				fmt.Printf("no deps")
			}
			fmt.Printf("\n\n")
		}
	}

	// and print module details
	fmt.Printf("MODULES:\n")
	for _, mi := range mods {
		fmt.Printf("  %s ", mi.name)
		if mi.version == "" {
			fmt.Printf("(no version)\n")
		} else {
			fmt.Printf("(%s)\n", mi.version)
		}
		for _, pi := range mi.packageInfos {
			fmt.Printf("    %s\n", pi.importPath)
		}
	}
}
