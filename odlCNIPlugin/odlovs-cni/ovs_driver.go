/*
 * Copyright (c) 2017 Kontron Canada Company and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package main

import (
    "errors"
    "fmt"
    "reflect"
    "sync"
    "time"

    log "github.com/Sirupsen/logrus"
    "github.com/socketplane/libovsdb"
)

const (
    // Default bridge name
    DefaultBridgeName = "ovsbrk8s"

    // Default port used to set the bridge controller
    DefaultControllerPort = 6653

    // Default port used to set OVS Manager
    DefaultManagerPort = 6640

    // ovsdb operations
    insertOpr = "insert"
    deleteOpr = "delete"
)

// OVS driver state
type OvsDriver struct {
    // OVS client
    ovsClient *libovsdb.OvsdbClient

    // Name of the OVS bridge
    OvsBridgeName string

    // OVSDB cache
    ovsdbCache map[string]map[string]libovsdb.Row

    // read/write lock for accessing the cache
    lock sync.RWMutex
}

// Create a new OVS driver with Unix socket
// deafult socket file path "/var/run/openvswitch/db.sock"
// Default bridge (br_k8s) will be created
func DefaultOvsDriver() *OvsDriver {
    return NewOvsDriver(DefaultBridgeName)
}

// Create a new OVS driver with Unix socket
// default socket file path "/var/run/openvswitch/db.sock"
func NewOvsDriver(bridgeName string) *OvsDriver {
    if bridgeName == "" {
        log.Fatal("Bridge could not be empty")
        return nil
    }
    ovsDriver := new(OvsDriver)
    // connect over a Unix socket:
    ovs, err := libovsdb.ConnectWithUnixSocket("/var/run/openvswitch/db.sock")
    if err != nil {
        log.Fatal("Failed to connect to ovsdb")
    }

    // Setup state
    ovsDriver.ovsClient = ovs
    ovsDriver.OvsBridgeName = bridgeName
    ovsDriver.ovsdbCache = make(map[string]map[string]libovsdb.Row)

    go func() {
        // Register for notifications
        ovs.Register(ovsDriver)

        // Populate initial state into cache
        initial, _ := ovs.MonitorAll("Open_vSwitch", "")
        ovsDriver.populateCache(*initial)
    }()

    // sleep the main thread so that Cache can be populated
    time.Sleep(1 * time.Second)

    // Create the default bridge instance
    err = ovsDriver.CreateBridge(ovsDriver.OvsBridgeName)
    if err != nil {
        log.Fatalf("Error creating bridge. Err: %v", err)
    }

    return ovsDriver
}

// Delete : Cleanup the ovsdb driver. delete the bridge we created.
func (self *OvsDriver) Delete() {
    if self.ovsClient != nil {
        self.DeleteBridge(self.OvsBridgeName)
        log.Infof("Deleting OVS bridge: %s", self.OvsBridgeName)
        (*self.ovsClient).Disconnect()
    }
}

// 
func cleanDefault() {
    // connect over a Unix socket:
    ovs, err := libovsdb.ConnectWithUnixSocket("/var/run/openvswitch/db.sock")
    if err != nil {
        log.Fatal("Failed to connect to ovsdb")
    }
    ovsDriver := new(OvsDriver)
    ovsDriver.ovsClient = ovs
    ovsDriver.OvsBridgeName = DefaultBridgeName
    ovsDriver.Delete()
    ovsDriver.diconnect()
}

// Disconnect the ovsDriver.
func (d *OvsDriver) diconnect() {
    if d.ovsClient != nil {
        (*d.ovsClient).Disconnect()
    }
}

// Populate local cache of ovs state
func (self *OvsDriver) populateCache(updates libovsdb.TableUpdates) {
    // lock the cache for write
    self.lock.Lock()
    defer self.lock.Unlock()

    for table, tableUpdate := range updates.Updates {
        if _, ok := self.ovsdbCache[table]; !ok {
            self.ovsdbCache[table] = make(map[string]libovsdb.Row)
        }

        for uuid, row := range tableUpdate.Rows {
            empty := libovsdb.Row{}
            if !reflect.DeepEqual(row.New, empty) {
                self.ovsdbCache[table][uuid] = row.New
            } else {
                delete(self.ovsdbCache[table], uuid)
            }
        }
    }
}

// Dump the contents of the cache into stdout
func (self *OvsDriver) PrintCache() {
    // lock the cache for read
    self.lock.RLock()
    defer self.lock.RUnlock()

    fmt.Printf("OvsDB Cache: \n")

    // walk the local cache
    for tName, table := range self.ovsdbCache {
        fmt.Printf("Table: %s\n", tName)
        for uuid, row := range table {
            fmt.Printf("  Row: UUID: %s\n", uuid)
            for fieldName, value := range row.Fields {
                fmt.Printf("    Field: %s, Value: %+v\n", fieldName, value)
            }
        }
    }
}

// Get the UUID for root
func (self *OvsDriver) getRootUuid() libovsdb.UUID {
    // lock the cache for read
    self.lock.RLock()
    defer self.lock.RUnlock()

    // find the matching uuid
    for uuid := range self.ovsdbCache["Open_vSwitch"] {
        return libovsdb.UUID{GoUUID: uuid}
    }
    return libovsdb.UUID{}
}

// Wrapper for ovsDB transaction
func (self *OvsDriver) ovsdbTransact(ops []libovsdb.Operation) error {
    // Print out what we are sending
    log.Debugf("Transaction: %+v\n", ops)

    // Perform OVSDB transaction
    reply, _ := self.ovsClient.Transact("Open_vSwitch", ops...)

    if len(reply) < len(ops) {
        log.Errorf("Unexpected number of replies. Expected: %d, Recvd: %d", len(ops), len(reply))
        return errors.New("OVS transaction failed. Unexpected number of replies")
    }

    // Parse reply and look for errors
    for i, o := range reply {
        if o.Error != "" && i < len(ops) {
            return errors.New("OVS Transaction failed err " + o.Error + "Details: " + o.Details)
        } else if o.Error != "" {
            return errors.New("OVS Transaction failed err " + o.Error + "Details: " + o.Details)
        }
    }

    // Return success
    return nil
}

// Create bridge to the ovs instance
func (self *OvsDriver) CreateBridge(bridgeName string) error {
    if self.IsBridgePresent(bridgeName) {
        return fmt.Errorf("Bridge %s already exist ", bridgeName)
    }

    namedUuidStr := "odlbridge"
    protocols := []string{"OpenFlow10", "OpenFlow11", "OpenFlow12", "OpenFlow13"}
    brOp := libovsdb.Operation{}
    bridge := make(map[string]interface{})
    bridge["name"] = bridgeName
    bridge["protocols"], _ = libovsdb.NewOvsSet(protocols)
    //bridge["fail_mode"] = "secure"
    brOp = libovsdb.Operation{
        Op:       insertOpr,
        Table:    "Bridge",
        Row:      bridge,
        UUIDName: namedUuidStr,
    }

    // mutating the open_vswitch table.
    brUuid := []libovsdb.UUID{{GoUUID: namedUuidStr}}
    mutateUuid := brUuid
    mutateSet, _ := libovsdb.NewOvsSet(mutateUuid)
    mutation := libovsdb.NewMutation("bridges", insertOpr, mutateSet)
    condition := libovsdb.NewCondition("_uuid", "==", self.getRootUuid())

    mutateOp := libovsdb.Operation{
        Op:        "mutate",
        Table:     "Open_vSwitch",
        Mutations: []interface{}{mutation},
        Where:     []interface{}{condition},
    }

    operations := []libovsdb.Operation{brOp, mutateOp}
    self.ovsdbTransact(operations)
    return self.CreatePort(bridgeName, "internal", 0)
}

// Delete a bridge from ovs instance
func (self *OvsDriver) DeleteBridge(bridgeName string) error {
    // lock the cache for read
    self.lock.RLock()
    defer self.lock.RUnlock()

    namedUuidStr := "odlbridge"
    brUuid := []libovsdb.UUID{{GoUUID: namedUuidStr}}

    brOp := libovsdb.Operation{}
    condition := libovsdb.NewCondition("name", "==", bridgeName)
    brOp = libovsdb.Operation{
        Op:    deleteOpr,
        Table: "Bridge",
        Where: []interface{}{condition},
    }

    // fetch the br-uuid from cache
    for uuid, row := range self.ovsdbCache["Bridge"] {
        name := row.Fields["name"].(string)
        if name == bridgeName {
            brUuid = []libovsdb.UUID{{GoUUID: uuid}}
            break
        }
    }

    mutateUuid := brUuid
    mutateSet, _ := libovsdb.NewOvsSet(mutateUuid)
    mutation := libovsdb.NewMutation("bridges", deleteOpr, mutateSet)
    condition = libovsdb.NewCondition("_uuid", "==", self.getRootUuid())

    mutateOp := libovsdb.Operation{
        Op:        "mutate",
        Table:     "Open_vSwitch",
        Mutations: []interface{}{mutation},
        Where:     []interface{}{condition},
    }

    operations := []libovsdb.Operation{brOp, mutateOp}
    return self.ovsdbTransact(operations)
}

// Create port in OVS bridge
func (self *OvsDriver) CreatePort(intfName string, intfType string, vlanTag uint) error {
    //check if port already created
    if self.IsPortNamePresent(intfName) {
        return fmt.Errorf("port %s already exist ", intfName)
    }
    portUuidStr := intfName
    intfUuidStr := "int"+intfName
    portUuid := []libovsdb.UUID{{GoUUID: portUuidStr}}
    intfUuid := []libovsdb.UUID{{GoUUID: intfUuidStr}}
    var err error = nil

    intf := make(map[string]interface{})
    intf["name"] = intfName
    intf["type"] = intfType

    // Add an entry in Interface table
    intfOp := libovsdb.Operation{
        Op:       insertOpr,
        Table:    "Interface",
        Row:      intf,
        UUIDName: intfUuidStr,
    }

    // insert row in Port table
    port := make(map[string]interface{})
    port["name"] = intfName
    if vlanTag != 0 {
        port["vlan_mode"] = "access"
        port["tag"] = vlanTag
    } else {
        port["vlan_mode"] = "trunk"
    }

    port["interfaces"], err = libovsdb.NewOvsSet(intfUuid)
    if err != nil {
        fmt.Println("error at interface uuid")
        return err
    }

    // Add an entry in Port table
    portOp := libovsdb.Operation{
        Op:       insertOpr,
        Table:    "Port",
        Row:      port,
        UUIDName: portUuidStr,
    }

    // mutate the Ports column in the Bridge table
    mutateSet, _ := libovsdb.NewOvsSet(portUuid)
    mutation := libovsdb.NewMutation("ports", insertOpr, mutateSet)
    condition := libovsdb.NewCondition("name", "==", self.OvsBridgeName)
    mutateOp := libovsdb.Operation{
        Op:        "mutate",
        Table:     "Bridge",
        Mutations: []interface{}{mutation},
        Where:     []interface{}{condition},
    }

    operations := []libovsdb.Operation{intfOp, portOp, mutateOp}
    return self.ovsdbTransact(operations)
}

// Delete a port from OVS bridge
func (self *OvsDriver) DeletePort(intfName string) error {
    // lock the cache for read
    self.lock.RLock()
    defer self.lock.RUnlock()

    portUuidStr := intfName
    portUuid := []libovsdb.UUID{{GoUUID: portUuidStr}}

    condition := libovsdb.NewCondition("name", "==", intfName)
    intfOp := libovsdb.Operation{
        Op:    deleteOpr,
        Table: "Interface",
        Where: []interface{}{condition},
    }

    condition = libovsdb.NewCondition("name", "==", intfName)
    portOp := libovsdb.Operation{
        Op:    deleteOpr,
        Table: "Port",
        Where: []interface{}{condition},
    }

    // fetch the port-uuid from cache
    for uuid, row := range self.ovsdbCache["Port"] {
        name := row.Fields["name"].(string)
        if name == intfName {
            portUuid = []libovsdb.UUID{{GoUUID: uuid}}
            break
        }
    }

    // mutate the Ports column of Bridge table
    mutateSet, _ := libovsdb.NewOvsSet(portUuid)
    mutation := libovsdb.NewMutation("ports", deleteOpr, mutateSet)
    condition = libovsdb.NewCondition("name", "==", self.OvsBridgeName)
    mutateOp := libovsdb.Operation{
        Op:        "mutate",
        Table:     "Bridge",
        Mutations: []interface{}{mutation},
        Where:     []interface{}{condition},
    }

    operations := []libovsdb.Operation{intfOp, portOp, mutateOp}
    return self.ovsdbTransact(operations)
}

// Create vtep port on the OVS bridge
func (self *OvsDriver) CreateVtep(intfName string, vtepRemoteIP string) error {
    portUuidStr := intfName
    intfUuidStr := fmt.Sprintf("Intf%s", intfName)
    portUuid := []libovsdb.UUID{{GoUUID: portUuidStr}}
    intfUuid := []libovsdb.UUID{{GoUUID: intfUuidStr}}
    intfType := "vxlan"
    var err error = nil

    intf := make(map[string]interface{})
    intf["name"] = intfName
    intf["type"] = intfType

    intfOptions := make(map[string]interface{})
    intfOptions["remote_ip"] = vtepRemoteIP
    intfOptions["key"] = "flow"

    intf["options"], err = libovsdb.NewOvsMap(intfOptions)
    if err != nil {
        log.Errorf("error '%s' creating options from %v \n", err, intfOptions)
        return err
    }

    // Add an entry in Interface table
    intfOp := libovsdb.Operation{
        Op:       insertOpr,
        Table:    "Interface",
        Row:      intf,
        UUIDName: intfUuidStr,
    }

    // insert/delete a row in Port table
    port := make(map[string]interface{})
    port["name"] = intfName
    port["vlan_mode"] = "trunk"
    port["interfaces"], err = libovsdb.NewOvsSet(intfUuid)
    if err != nil {
        return err
    }

    // Add an entry in Port table
    portOp := libovsdb.Operation{
        Op:       insertOpr,
        Table:    "Port",
        Row:      port,
        UUIDName: portUuidStr,
    }

    // mutate the Ports column of the row in the Bridge table
    mutateSet, _ := libovsdb.NewOvsSet(portUuid)
    mutation := libovsdb.NewMutation("ports", insertOpr, mutateSet)
    condition := libovsdb.NewCondition("name", "==", self.OvsBridgeName)
    mutateOp := libovsdb.Operation{
        Op:        "mutate",
        Table:     "Bridge",
        Mutations: []interface{}{mutation},
        Where:     []interface{}{condition},
    }

    // Execute the transaction
    operations := []libovsdb.Operation{intfOp, portOp, mutateOp}
    return self.ovsdbTransact(operations)
}

// Add controller to bridge
// if portNo is 0 default port 6653 will be used
func (self *OvsDriver) SetActiveController(ipAddress string, portNo uint16) error {
    if ipAddress == "" {
        return errors.New("IP address cannot be empty")
    }
    if portNo == 0 {
        portNo = DefaultControllerPort
    }
    target := fmt.Sprintf("tcp:%s:%d", ipAddress, portNo)
    return self.SetController(target)
}

// Add passive controller to bridge
// if portNo is 0 default port 6653 will be used
func (self *OvsDriver) SetPassiveController(portNo uint16) error {
    if portNo == 0 {
        portNo = DefaultControllerPort
    }
    target := fmt.Sprintf("ptcp:%d", portNo)
    return self.SetController(target)
}

// Add controller configuration to OVS
// target should contain ipAddress and port ex tcp:127.0.0.1:6653
func (self *OvsDriver) SetController(target string) error {
    if target == "" {
        return errors.New("target cannot be empty")
    }
    ctrlerUuidStr := fmt.Sprintf("local")
    ctrlerUuid := []libovsdb.UUID{{GoUUID: ctrlerUuidStr}}

    // If controller already exists, nothing to do
    if self.IsControllerPresent(target) {
        return errors.New(fmt.Sprintf("Controller %s already exist", target))
    }

    // insert a row in Controller table
    controller := make(map[string]interface{})
    controller["target"] = target

    // Add an entry in Controller table
    ctrlerOp := libovsdb.Operation{
        Op:       insertOpr,
        Table:    "Controller",
        Row:      controller,
        UUIDName: ctrlerUuidStr,
    }

    // mutate the Controller column of Bridge table
    mutateSet, _ := libovsdb.NewOvsSet(ctrlerUuid)
    mutation := libovsdb.NewMutation("controller", insertOpr, mutateSet)
    condition := libovsdb.NewCondition("name", "==", self.OvsBridgeName)
    mutateOp := libovsdb.Operation{
        Op:        "mutate",
        Table:     "Bridge",
        Mutations: []interface{}{mutation},
        Where:     []interface{}{condition},
    }

    operations := []libovsdb.Operation{ctrlerOp, mutateOp}
    return self.ovsdbTransact(operations)
}

func (self *OvsDriver) DeleteController(target string) error {
    // FIXME:
    return nil
}

//
func (self *OvsDriver) SetActiveManager(ipAddress string, portNo int) error {
    if ipAddress == "" {
        return errors.New("IP address cannot be empty")
    }
    if portNo == 0 {
        portNo = DefaultControllerPort
    }
    target := fmt.Sprintf("tcp:%s:%d", ipAddress, portNo)
    return self.SetManager(target)
}

//
func (self *OvsDriver) SetPassiveManager(portNo int) error {
    if portNo == 0 {
        portNo = DefaultControllerPort
    }
    target := fmt.Sprintf("tcp:%d", portNo)
    return self.SetManager(target)
}

//
func (self *OvsDriver) SetManager(target string) error {
    // FIXME:
    return nil
}

//
func (self *OvsDriver) DeleteManager(target string) error {
    // FIXME:
    return nil
}

// Check if VTEP already exists
func (self *OvsDriver) IsVtepPresent(remoteIP string) (bool, string) {
    self.lock.RLock()
    defer self.lock.RUnlock()

    for tName, table := range self.ovsdbCache {
        if tName == "Interface" {
            for _, row := range table {
                options := row.Fields["options"]
                switch optMap := options.(type) {
                case libovsdb.OvsMap:
                    if optMap.GoMap["remote_ip"] == remoteIP {
                        value := row.Fields["name"]
                        switch t := value.(type) {
                        case string:
                            return true, t
                        default:
                            return false, ""
                        }
                    }
                default:
                    return false, ""
                }
            }
        }
    }

    return false, ""
}


// Check the local cache and see if the portname is taken already
// HACK alert: This is used to pick next port number instead of managing
// port number space actively across agent restarts
func (self *OvsDriver) IsPortNamePresent(intfName string) bool {
    self.lock.RLock()
    defer self.lock.RUnlock()

    for tName, table := range self.ovsdbCache {
        if tName == "Port" {
            for _, row := range table {
                for fieldName, value := range row.Fields {
                    if fieldName == "name" {
                        if value == intfName {
                            // Interface name exists.
                            return true
                        }
                    }
                }
            }
        }
    }

    return false
}

// Check if the bridge already exists
func (self *OvsDriver) IsBridgePresent(bridgeName string) bool {
    self.lock.RLock()
    defer self.lock.RUnlock()

    for tName, table := range self.ovsdbCache {
        if tName == "Bridge" {
            for _, row := range table {
                for fieldName, value := range row.Fields {
                    if fieldName == "name" {
                        if value == bridgeName {
                            // Interface name exists.
                            return true
                        }
                    }
                }
            }
        }
    }

    return false
}

// Check if Controller already exists
func (self *OvsDriver) IsControllerPresent(target string) bool {
    self.lock.RLock()
    defer self.lock.RUnlock()

    for tName, table := range self.ovsdbCache {
        if tName == "Controller" {
            for _, row := range table {
                for fieldName, value := range row.Fields {
                    if fieldName == "target" {
                        if value == target {
                            // Controller exists.
                            return true
                        }
                    }
                }
            }
        }
    }

    return false
}

// Return OFP port number for an interface
func (self *OvsDriver) GetOfpPortNo(intfName string) (uint32, error) {
    retryNo := 0
    condition := libovsdb.NewCondition("name", "==", intfName)
    selectOp := libovsdb.Operation{
        Op:    "select",
        Table: "Interface",
        Where: []interface{}{condition},
    }

    for {
        row, err := self.ovsClient.Transact("Open_vSwitch", selectOp)

        if err == nil && len(row) > 0 && len(row[0].Rows) > 0 {
            value := row[0].Rows[0]["ofport"]
            if reflect.TypeOf(value).Kind() == reflect.Float64 {
                //retry few more time. Due to asynchronous call between
                //port creation and populating ovsdb entry for the interface.
                var ofpPort uint32 = uint32(reflect.ValueOf(value).Float())
                return ofpPort, nil
            }
        }
        time.Sleep(200 * time.Millisecond)

        if retryNo == 5 {
            return 0, errors.New("ofPort not found")
        }
        retryNo++
    }
}

// ************************ Notification handler for OVS DB changes ****************
func (self *OvsDriver) Update(context interface{}, tableUpdates libovsdb.TableUpdates) {
    self.populateCache(tableUpdates)
}
func (self *OvsDriver) Disconnected(ovsClient *libovsdb.OvsdbClient) {
    log.Infof("OVS Driver disconnected")
}
func (self *OvsDriver) Locked([]interface{}) {
}
func (self *OvsDriver) Stolen([]interface{}) {
}
func (self *OvsDriver) Echo([]interface{}) {
}