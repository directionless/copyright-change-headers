package main

//go:generate go-bindata -nometadata -nocompress -pkg internal -o internal/old-licenses.go internal/old-licenses/

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/directionless/copyright-change-repo-headers/cmd/osquery-change/internal"
	"github.com/directionless/copyright-change-repo-headers/pkg/changer"
	"github.com/peterbourgon/ff"
)

var baseLicense = []string{
	"Copyright (c) 2014-present, The osquery authors",
	"",
	"This source code is licensed as defined by the LICENSE file found in the",
	"root directory of this source tree.",
	"",
	"SPDX-License-Identifier: (Apache-2.0 OR GPL-2.0-only)",
}

func main() {
	fs := flag.NewFlagSet("osquery-copyright-change", flag.ExitOnError)
	var (
		flCheckoutDir = fs.String("checkout", "", "Where's your osquery checkout")
	)

	checkError(ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileParser(ff.PlainParser),
	))

	if *flCheckoutDir == "" {
		fmt.Println(`Missing required "checkout" option`)
		os.Exit(1)
	}

	c := changer.New(baseLicense,

		// Normalize years, to make the various licenses we have simpler.
		changer.WithRegexCleaner(regexp.MustCompile(`Copyright \(c\) 20(.*)Facebook, Inc.`),
			[]byte(`Copyright (c) 2014-present, Facebook, Inc.`)),

		changer.WithOldLicense(changer.CStyle, internal.MustAsset("internal/old-licenses/c1")),
		changer.WithOldLicense(changer.CStyle, internal.MustAsset("internal/old-licenses/c2")),
		changer.WithOldLicense(changer.CStyle, internal.MustAsset("internal/old-licenses/c3")),
		changer.WithOldLicense(changer.CStyle, internal.MustAsset("internal/old-licenses/c4")),
		changer.WithOldLicense(changer.CStyle, internal.MustAsset("internal/old-licenses/c5")),
		changer.WithOldLicense(changer.CStyle, internal.MustAsset("internal/old-licenses/c6")),
		changer.WithOldLicense(changer.CStyle, internal.MustAsset("internal/old-licenses/c7")),
		changer.WithOldLicense(changer.CStyle, internal.MustAsset("internal/old-licenses/c8")),

		changer.WithOldLicense(changer.ShStyle, internal.MustAsset("internal/old-licenses/sh1")),
		changer.WithOldLicense(changer.ShStyle, internal.MustAsset("internal/old-licenses/sh2")),
		changer.WithOldLicense(changer.ShStyle, internal.MustAsset("internal/old-licenses/sh3")),

		changer.WithOldLicense(changer.PyStyle, internal.MustAsset("internal/old-licenses/sh1")),
		changer.WithOldLicense(changer.PyStyle, internal.MustAsset("internal/old-licenses/sh2")),
		changer.WithOldLicense(changer.PyStyle, internal.MustAsset("internal/old-licenses/sh3")),
	)
	checkError(c.Walk(*flCheckoutDir))

	c.Status(os.Stdout)
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Got Error: %v\n", err)
		os.Exit(1)
	}
}
