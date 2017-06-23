/*
 * Copyright © 2017 Copyright c 2017 Ericsson India Global Services Pvt Ltd. and others.All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */
package org.opendaylight.coe.southbound.impl;

import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;
import javax.inject.Inject;
import org.opendaylight.controller.md.sal.binding.api.DataBroker;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class CoeSouthboundProvider {

    private static final Logger LOG = LoggerFactory.getLogger(CoeSouthboundProvider.class);

    private final DataBroker dataBroker;

    @Inject
    public CoeSouthboundProvider(final DataBroker dataBroker) {
        this.dataBroker = dataBroker;
    }

    @PostConstruct
    public void init() {
        LOG.info("CoeSouthboundProvider Session Initiated");
    }

    @PreDestroy
    public void close() {
        LOG.info("CoeSouthboundProvider Closed");
    }
}