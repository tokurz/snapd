// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (c) 2016-2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more dtails.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.
 *
 */

package builtin

import (
	"strings"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/seccomp"
)

const mirSummary = `allows operating as the Mir server`

const mirBaseDeclarationSlots = `
  mir:
    allow-installation:
      slot-snap-type:
        - app
    deny-connection: true
`

const mirPermanentSlotAppArmor = `
# Description: Allow operating as the Mir server. This gives privileged access
# to the system.

# needed since Mir is the display server, to configure tty devices
capability sys_tty_config,
/dev/tty[0-9]* rw,

/{dev,run}/shm/\#* mrw,
/run/mir_socket rw,

# Needed for mode setting via drmSetMaster() and drmDropMaster()
capability sys_admin,

# NOTE: this allows reading and inserting all input events
/dev/input/* rw,

# For using udev
network netlink raw,
/run/udev/data/c13:[0-9]* r,
/run/udev/data/+input:input[0-9]* r,
/run/udev/data/+platform:* r,
`

const mirPermanentSlotSecComp = `
# Description: Allow operating as the mir server. This gives privileged access
# to the system.
# Needed for server launch
bind
listen
# Needed by server upon client connect
accept
accept4
shmctl
# for udev
socket AF_NETLINK - NETLINK_KOBJECT_UEVENT
`

const mirConnectedSlotAppArmor = `
# Description: Permit clients to use Mir
unix (receive, send) type=seqpacket addr=none peer=(label=###PLUG_SECURITY_TAGS###),
`

const mirConnectedPlugAppArmor = `
# Description: Permit clients to use Mir
unix (receive, send) type=seqpacket addr=none peer=(label=###SLOT_SECURITY_TAGS###),
/run/mir_socket rw,
/run/user/[0-9]*/mir_socket rw,

# Lttng tracing is very noisy and should not be allowed by confined apps. Can
# safely deny. LP: #1260491
deny /{dev,run,var/run}/shm/lttng-ust-* rw,
`

type mirInterface struct{}

func (iface *mirInterface) Name() string {
	return "mir"
}

func (iface *mirInterface) MetaData() interfaces.MetaData {
	return interfaces.MetaData{
		Summary:              mirSummary,
		BaseDeclarationSlots: mirBaseDeclarationSlots,
	}
}

func (iface *mirInterface) AppArmorConnectedPlug(spec *apparmor.Specification, plug *interfaces.Plug, plugAttrs map[string]interface{}, slot *interfaces.Slot, slotAttrs map[string]interface{}) error {
	old := "###SLOT_SECURITY_TAGS###"
	new := slotAppLabelExpr(slot)
	snippet := strings.Replace(mirConnectedPlugAppArmor, old, new, -1)
	spec.AddSnippet(snippet)
	return nil
}

func (iface *mirInterface) AppArmorConnectedSlot(spec *apparmor.Specification, plug *interfaces.Plug, plugAttrs map[string]interface{}, slot *interfaces.Slot, slotAttrs map[string]interface{}) error {
	old := "###PLUG_SECURITY_TAGS###"
	new := plugAppLabelExpr(plug)
	snippet := strings.Replace(mirConnectedSlotAppArmor, old, new, -1)
	spec.AddSnippet(snippet)
	return nil
}

func (iface *mirInterface) AppArmorPermanentSlot(spec *apparmor.Specification, slot *interfaces.Slot) error {
	spec.AddSnippet(mirPermanentSlotAppArmor)
	return nil
}

func (iface *mirInterface) SecCompPermanentSlot(spec *seccomp.Specification, slot *interfaces.Slot) error {
	spec.AddSnippet(mirPermanentSlotSecComp)
	return nil
}

func (iface *mirInterface) SanitizePlug(plug *interfaces.Plug) error {
	return nil
}

func (iface *mirInterface) SanitizeSlot(slot *interfaces.Slot) error {
	return nil
}

func (iface *mirInterface) AutoConnect(*interfaces.Plug, *interfaces.Slot) bool {
	return true
}

func init() {
	registerIface(&mirInterface{})
}
