/*
 * Copyright (c) 2017 Kontron - S & T Company and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package main

import (
    "flag"
    "fmt"
    "github.com/containernetworking/cni/pkg/version"
    "github.com/containernetworking/cni/pkg/skel"
)

const APP_VERSION = "0.1"

func cmdAdd(args *skel.CmdArgs) error {
    fmt.Println("Add Cmd.")
    return nil
}

func cmdDel(args *skel.CmdArgs) error {
    fmt.Println("Add Cmd.")
    return nil
}

func main() {
    flag.Parse() // Scan the arguments list 

    skel.PluginMain(cmdAdd, cmdDel, version.All)
}

