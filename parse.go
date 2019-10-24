// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"log"
	"strings"
)

// parses the output of the go list command into package info

type packageInfo struct {
	isStandard    bool
	importPath    string
	modName       string
	modVersion    string
	deps          []string
	isMainPackage bool
}

type moduleInfo struct {
	name         string
	version      string
	packageInfos []*packageInfo
}

// parses one line of output into a packageInfo record
func parsePackgeInfo(s string) (*packageInfo, error) {
	// trim leading and trailing apostrophes
	s = strings.TrimPrefix(s, "'")
	s = strings.TrimSuffix(s, "'")

	// break into fields
	fields := strings.Split(s, "#")
	if len(fields) != 4 {
		return nil, fmt.Errorf("cannot parse line: %s", s)
	}

	pi := &packageInfo{}

	switch fields[0] {
	case "true":
		pi.isStandard = true
	case "false":
		pi.isStandard = false
	default:
		return nil, fmt.Errorf("invalid result for Standard: %s", fields[0])
	}

	pi.importPath = fields[1]

	if fields[2] != "<nil>" {
		// check whether there is a space followed by version info
		modFields := strings.Split(fields[2], " ")
		pi.modName = modFields[0]
		if len(modFields) > 1 {
			pi.modVersion = modFields[1]
		}
	}

	pi.deps = parseDepsString(fields[3])

	return pi, nil
}

// turn string with list of deps into a []string
func parseDepsString(s string) []string {
	// trim leading and trailing brackets
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")

	// now split by space separators
	depStrs := strings.Split(s, " ")

	// filter out empty strings
	deps := []string{}
	for _, d := range depStrs {
		if d != "" {
			deps = append(deps, d)
		}
	}

	return deps
}

// parse output from a go list command into a mapping of
// package names to packageInfo records
func parseGoListOutput(out []byte, mainPackage string) (map[string]*packageInfo, error) {
	lines := strings.Split(string(out), "\n")
	pis := map[string]*packageInfo{}
	for _, l := range lines {
		if l == "" {
			break
		}
		pi, err := parsePackgeInfo(l)
		if err != nil {
			return nil, err
		}
		pis[pi.importPath] = pi
	}

	// also update the main package
	mainPi := pis[mainPackage]
	mainPi.isMainPackage = true

	return pis, nil
}

// gather collection of packageInfos into set of modules
func getModules(pis map[string]*packageInfo) map[string]*moduleInfo {
	mods := map[string]*moduleInfo{}

	for _, pi := range pis {
		// ignore standard packages
		if pi.isStandard {
			continue
		}
		// ignore packages with no module
		if pi.modName == "" {
			continue
		}
		mod, ok := mods[pi.modName]
		if ok {
			// confirm module version is the same as previously seen
			// if it isn't, something is wrong, i think?
			if pi.modVersion != mod.version {
				log.Fatalf("error: for module %s, previously saw version %s but now seeing version %s; bailing", mod.name, mod.version, pi.modVersion)
			}
		} else {
			// create a new entry for this module
			mod = &moduleInfo{
				name:         pi.modName,
				version:      pi.modVersion,
				packageInfos: []*packageInfo{},
			}
			mods[pi.modName] = mod
		}
		mod.packageInfos = append(mod.packageInfos, pi)
	}

	return mods
}
