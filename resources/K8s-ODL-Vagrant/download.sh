#!/bin/sh
#mvn dependency:get -DrepoUrl="https://nexus.opendaylight.org/content/repositories/opendaylight.snapshot/" -DgroupId=org.opendaylight.integration -DartifactId=karaf -Dversion=0.9.0-SNAPSHOT -Dpackaging=tar.gz
#cp $HOME/.m2/repository/org/opendaylight/integration/karaf/0.9.0-SNAPSHOT/karaf-0.9.0-SNAPSHOT.tar.gz .
wget -c "https://dl.google.com/go/go1.10.linux-amd64.tar.gz"
rm karaf-0.9.0-SNAPSHOT.tar.gz
curl -o karaf-0.9.0-SNAPSHOT.tar.gz -J -L "https://nexus.opendaylight.org/service/local/artifact/maven/content?r=opendaylight.snapshot&g=org.opendaylight.integration&a=karaf&e=tar.gz&v=0.9.0-SNAPSHOT"
