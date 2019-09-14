/*
 * Copyright © 2017 Ericsson India Global Services Pvt Ltd. and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */
package org.opendaylight.coe.cli.commands;

import org.apache.karaf.shell.commands.Command;
import org.apache.karaf.shell.console.OsgiCommandSupport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * This is an example class. The class name can be renamed to match the command implementation that it will invoke.
 * Specify command details by updating the fields in the Command annotation below.
 */
@Command(scope = "coe", name = "coe", description = "add a description for the command")
public class CoeCliTestCommand extends OsgiCommandSupport {

    private static final Logger LOG = LoggerFactory.getLogger(CoeCliTestCommand.class);

    @Override
    protected Object doExecute() throws Exception {
        //TODO will come in subsequent patches
        return null;
    }
}
