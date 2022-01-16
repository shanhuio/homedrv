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
	"shanhu.io/homedrv/drvapi"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/homedrv/homeapp"
	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/dock"
)

type ncfront struct {
	core homeapp.Core
}

func newNCFront(c homeapp.Core) *ncfront { return &ncfront{core: c} }

func (n *ncfront) cont() *dock.Cont {
	return dock.NewCont(n.core.Docker(), appCont(n.core, nameNCFront))
}

func (n *ncfront) createCont(image string) (*dock.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}

	nextcloudAddr := appCont(n.core, nameNextcloud) + ":80"
	config := &dock.ContConfig{
		Name:          appCont(n.core, nameNCFront),
		Network:       appNetwork(n.core),
		Env:           map[string]string{"NEXTCLOUD": nextcloudAddr},
		AutoRestart:   true,
		JSONLogConfig: dock.LimitedJSONLog(),
		Labels:        drvcfg.NewNameLabel(nameNCFront),
	}
	return dock.CreateCont(n.core.Docker(), image, config)
}

func (n *ncfront) startWithImage(image string) error {
	cont, err := n.createCont(image)
	if err != nil {
		return errcode.Annotate(err, "create ncfront container")
	}
	return cont.Start()
}

func (n *ncfront) install(image string) error {
	return n.startWithImage(image)
}

func (n *ncfront) update(image string) error {
	cont := appCont(n.core, nameNCFront)
	if err := dropContIfDifferent(n.core.Docker(), cont, image); err != nil {
		if err == errSameImage {
			return nil
		}
		return err
	}
	return n.startWithImage(image)
}

func (n *ncfront) Start() error { return n.cont().Start() }
func (n *ncfront) Stop() error  { return n.cont().Stop() }

func (n *ncfront) Change(from, to *drvapi.AppMeta) error {
	if from != nil {
		if err := n.cont().Drop(); err != nil {
			return errcode.Annotate(err, "drop old ncfront container")
		}
	}
	if to == nil {
		return nil
	}
	return n.install(appImage(to))
}
