/*
 * Copyright © 2017 Copyright c 2017 Ericsson India Global Services Pvt Ltd. and others.All rights reserved. and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */
package org.opendaylight.coe.northbound.impl;

import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;
import javax.inject.Inject;
import org.opendaylight.controller.md.sal.binding.api.DataBroker;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;


public class CoeNorthboundProvider {

    private static final Logger LOG = LoggerFactory.getLogger(CoeNorthboundProvider.class);

    private final DataBroker dataBroker;

    @Inject
    public CoeNorthboundProvider(final DataBroker dataBroker) {
        this.dataBroker = dataBroker;
    }

    @PostConstruct
    public void init() {
        LOG.info("CoeNorthboundProvider Session Initiated");
    }

    @PreDestroy
    public void close() {
        LOG.info("CoeNorthboundProvider Closed");
    }
}