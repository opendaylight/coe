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
	log "github.com/Sirupsen/logrus"
	"github.com/socketplane/libovsdb"
	"reflect"
	"sync"
	"time"
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
	mutateOpr = "mutate"
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
// Default bridge (ovsbrk8s) will be created
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
		ovsDriver.PopulateCache(*initial)
	}()

	// sleep the main thread so that Cache can be populated
	time.Sleep(1 * time.Second)

	// Create the default bridge instance
	ovsDriver.CreateBridge(ovsDriver.OvsBridgeName)

	return ovsDriver
}

// Delete the bridge we created and disconnect the ovsDriver.
func (self *OvsDriver) Delete() {
	if self.ovsClient != nil {
		self.DeleteBridge(self.OvsBridgeName)
		log.Infof("Deleting OVS bridge: %s", self.OvsBridgeName)
		(*self.ovsClient).Disconnect()
	}
}

// Disconnect the ovsDriver.
func (self *OvsDriver) Disconnect() {
	if self.ovsClient != nil {
		(*self.ovsClient).Disconnect()
	}
}

// Populate local cache of ovs state
func (self *OvsDriver) PopulateCache(updates libovsdb.TableUpdates) {
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

// Get the UUID for root
func (self *OvsDriver) GetRootUuid() libovsdb.UUID {
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
func (self *OvsDriver) OvsdbTransact(ops []libovsdb.Operation) error {
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
		return nil
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
	condition := libovsdb.NewCondition("_uuid", "==", self.GetRootUuid())

	mutateOp := libovsdb.Operation{
		Op:        mutateOpr,
		Table:     "Open_vSwitch",
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{brOp, mutateOp}
	err := self.OvsdbTransact(operations)
	if err != nil {
		return fmt.Errorf("Error while creating ovs bridge %v", err)
	}
	return self.CreatePort(bridgeName, "internal", 0, "", "")
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
	condition = libovsdb.NewCondition("_uuid", "==", self.GetRootUuid())

	mutateOp := libovsdb.Operation{
		Op:        mutateOpr,
		Table:     "Open_vSwitch",
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{brOp, mutateOp}
	return self.OvsdbTransact(operations)
}

// Create port in OVS bridge
func (self *OvsDriver) CreatePort(intfName string, intfType string, vlanTag uint, externalId string, intfMacAddress string) error {
	//check if port already created
	if self.IsPortNamePresent(intfName) {
		return nil
	}
	portUuidStr := intfName
	intfUuidStr := "int" + intfName
	portUuid := []libovsdb.UUID{{GoUUID: portUuidStr}}
	intfUuid := []libovsdb.UUID{{GoUUID: intfUuidStr}}
	var err error = nil

	intf := make(map[string]interface{})
	intf["name"] = intfName
	if intfType != "" {
		intf["type"] = intfType
	}
	if externalId != "" {
		extIDs := make(map[string]string)
		extIDs["iface-id"] = externalId
		extIDs["attached-mac"] = intfMacAddress
		intf["external_ids"], _ = libovsdb.NewOvsMap(extIDs)
	}

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
	if externalId != "" {
		extIDs := make(map[string]string)
		extIDs["iface-id"] = externalId
		extIDs["attached-mac"] = intfMacAddress
		port["external_ids"], _ = libovsdb.NewOvsMap(extIDs)
	}
	if vlanTag != 0 {
		port["vlan_mode"] = "access"
		port["tag"] = vlanTag
	} else {
		port["vlan_mode"] = "trunk"
	}

	port["interfaces"], err = libovsdb.NewOvsSet(intfUuid)
	if err != nil {
		return fmt.Errorf("Error at interface uuid , %v", err)
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
		Op:        mutateOpr,
		Table:     "Bridge",
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{intfOp, portOp, mutateOp}
	return self.OvsdbTransact(operations)
}

// Delete port from OVS bridge By Name
func (self *OvsDriver) DeletePortByName(intfName string) error {
	if intfName == "" {
		return fmt.Errorf("intf Name could not be empty")
	}
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
		Op:        mutateOpr,
		Table:     "Bridge",
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{intfOp, portOp, mutateOp}
	return self.OvsdbTransact(operations)
}

// Add controller to OVSDriver bridge
// should contain ipAddress and port ex: 127.0.0.1 and 6653
// if portNo is 0 default port 6653 will be used
func (self *OvsDriver) SetActiveController(ipAddress string, portNo int) error {
	if ipAddress == "" {
		return errors.New("IP address cannot be empty")
	}
	if portNo == 0 {
		portNo = DefaultControllerPort
	}
	target := fmt.Sprintf("tcp:%s:%d", ipAddress, portNo)
	return self.SetController(target)
}

// Add passive controller to OVSDriver bridge
// if portNo is 0 default port 6653 will be used
func (self *OvsDriver) SetPassiveController(portNo int) error {
	if portNo == 0 {
		portNo = DefaultControllerPort
	}
	target := fmt.Sprintf("ptcp:%d", portNo)
	return self.SetController(target)
}

// Add controller configuration to OVSDriver bridge
func (self *OvsDriver) SetController(target string) error {
	if target == "" {
		return fmt.Errorf("target cannot be empty")
	}
	ctrlerUuidStr := fmt.Sprintf("local")
	ctrlerUuid := []libovsdb.UUID{{GoUUID: ctrlerUuidStr}}

	// If controller already exists, nothing to do
	if self.IsControllerPresent(target) {
		return fmt.Errorf("Controller %s already exist", target)
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
		Op:        mutateOpr,
		Table:     "Bridge",
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{ctrlerOp, mutateOp}
	return self.OvsdbTransact(operations)
}

// Add Manager Config to OVS
// should contain ipAddress and port ex: 127.0.0.1 and 6640
// if portNo is 0 default port 6640 will be used
func (self *OvsDriver) SetActiveManager(ipAddress string, portNo int) error {
	if ipAddress == "" {
		return errors.New("IP address cannot be empty")
	}
	if portNo == 0 {
		portNo = DefaultManagerPort
	}
	target := fmt.Sprintf("tcp:%s:%d", ipAddress, portNo)
	return self.SetManager(target)
}

// Add Manager Config to OVS
// if portNo is 0 default port 6640 will be used
func (self *OvsDriver) SetPassiveManager(portNo int) error {
	if portNo == 0 {
		portNo = DefaultManagerPort
	}
	target := fmt.Sprintf("ptcp:%d", portNo)
	return self.SetManager(target)
}

// Set the Manager Config to OVS
func (self *OvsDriver) SetManager(target string) error {
	if target == "" {
		return fmt.Errorf("target cannot be empty")
	}
	// If manager already exists, nothing to do
	if self.IsManagerPresent(target) {
		return fmt.Errorf("Manager %s already exist", target)
	}

	// insert a row in manager table
	managerUuidStr := fmt.Sprintf("odlmngr")
	manager := make(map[string]interface{})
	manager["target"] = target

	// Add an entry in Manager table
	managerOp := libovsdb.Operation{
		Op:       insertOpr,
		Table:    "Manager",
		Row:      manager,
		UUIDName: managerUuidStr,
	}

	// mutating the open_vswitch table.
	managerUuid := []libovsdb.UUID{{GoUUID: managerUuidStr}}
	mutateUuid := managerUuid
	mutateSet, _ := libovsdb.NewOvsSet(mutateUuid)
	mutation := libovsdb.NewMutation("manager_options", insertOpr, mutateSet)
	condition := libovsdb.NewCondition("_uuid", "==", self.GetRootUuid())

	mutateOp := libovsdb.Operation{
		Op:        mutateOpr,
		Table:     "Open_vSwitch",
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{managerOp, mutateOp}
	return self.OvsdbTransact(operations)
}

// Check the local cache and see if the portname is exist
func (self *OvsDriver) IsPortNamePresent(intfName string) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()
	for _, row := range self.ovsdbCache["Port"] {
		if name, ok := row.Fields["name"]; ok && name == intfName {
			return true
		}
	}
	return false
}

// Get Port Name by externalId
func (self *OvsDriver) GetPortNameByExternalId(externalId string) string {
	self.lock.RLock()
	defer self.lock.RUnlock()
	for _, row := range self.ovsdbCache["Port"] {
		if extIDs, ok := row.Fields["external_ids"]; ok {
			extIDsMap := extIDs.(libovsdb.OvsMap).GoMap
			if ifaceId, ok := extIDsMap["iface-id"]; ok && ifaceId == externalId {
				return row.Fields["name"].(string)
			}
		}
	}
	return ""
}

// Check if the bridge already exists
func (self *OvsDriver) IsBridgePresent(bridgeName string) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, row := range self.ovsdbCache["Bridge"] {
		if name, ok := row.Fields["name"]; ok && name == bridgeName {
			return true
		}
	}
	return false
}

// Check if Controller already exists
func (self *OvsDriver) IsControllerPresent(target string) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, row := range self.ovsdbCache["Controller"] {
		if ctlr, ok := row.Fields["target"]; ok && ctlr == target {
			return true
		}
	}
	return false
}

// Check if Manager already exists
func (self *OvsDriver) IsManagerPresent(target string) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, row := range self.ovsdbCache["Manager"] {
		if mangr, ok := row.Fields["target"]; ok && mangr == target {
			return true
		}
	}
	return false
}

// ************************ Notification handler for OVS DB changes ****************
func (self *OvsDriver) Update(context interface{}, tableUpdates libovsdb.TableUpdates) {
	self.PopulateCache(tableUpdates)
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