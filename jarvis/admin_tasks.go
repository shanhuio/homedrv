// Copyright (C) 2022  Shanhu Tech Inc.
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the
// Free Software Foundation, either version 3 of the License, or (at your
// option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package jarvis

import (
	"fmt"

	"shanhu.io/aries"
	"shanhu.io/homedrv/homeapp/nextcloud"
	"shanhu.io/misc/errcode"
)

type adminTasks struct {
	server *server
}

func (s *adminTasks) apiUpdate(c *aries.C, sig bool) error {
	s.server.updateSignal <- sig
	return nil
}

func (s *adminTasks) apiRecreateDoorway(c *aries.C) error {
	d := s.server.drive
	t := &taskRecreateDoorway{drive: d}
	return d.tasks.run("recreate doorway", t)
}

func (s *adminTasks) apiSetRootPassword(c *aries.C, pwd string) error {
	return s.server.users.setPassword(rootUser, pwd, nil)
}

func (s *adminTasks) apiDisableTOTP(c *aries.C, user string) error {
	return s.server.users.disableTOTP(user)
}

func (s *adminTasks) apiSetAPIKey(c *aries.C, keyBytes []byte) error {
	return s.server.keyRegistry.apiSet(c, keyBytes)
}

type taskReinstallApp struct {
	drive *drive
	name  string
}

func (t *taskReinstallApp) run() error {
	return t.drive.apps.reinstall(t.name)
}

func (s *adminTasks) apiSetNextcloudDataMount(c *aries.C, m string) error {
	d := s.server.drive
	if err := d.settings.Set(nextcloud.KeyDataMount, m); err != nil {
		return errcode.Annotate(err, "set nextcloud data mount")
	}
	t := &taskReinstallApp{drive: d, name: nextcloud.Name}
	return d.tasks.run("restart nextcloud", t)
}

func (s *adminTasks) apiSetNextcloudExtraMounts(
	c *aries.C, m map[string]string,
) error {
	d := s.server.drive
	if err := d.settings.Set(nextcloud.KeyExtraMounts, m); err != nil {
		return errcode.Annotate(err, "set nextcloud extra mounts")
	}
	t := &taskReinstallApp{drive: d, name: nextcloud.Name}
	return d.tasks.run("restart nextcloud", t)
}

func (s *adminTasks) apiReinstallApp(c *aries.C, name string) error {
	d := s.server.drive
	t := &taskReinstallApp{drive: d, name: name}
	return d.tasks.run(fmt.Sprintf("reinstall %s", name), t)
}

func adminTasksAPI(s *server) *aries.Router {
	tasks := &adminTasks{server: s}

	r := aries.NewRouter()
	r.Call("update", tasks.apiUpdate)
	r.Call("recreate-doorway", tasks.apiRecreateDoorway)
	r.Call("set-root-password", tasks.apiSetRootPassword)
	r.Call("disable-totp", tasks.apiDisableTOTP)
	r.Call("set-api-key", tasks.apiSetAPIKey)
	r.Call("reinstall-app", tasks.apiReinstallApp)
	r.Call("set-nextcloud-datamnt", tasks.apiSetNextcloudDataMount)
	r.Call("set-nextcloud-extramnt", tasks.apiSetNextcloudExtraMounts)
	return r
}
