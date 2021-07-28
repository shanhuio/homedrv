// Copyright (C) 2021  Shanhu Tech Inc.
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
	"net/url"
	"strings"
	"sync"

	"shanhu.io/aries/creds"
	"shanhu.io/homedrv/burmilla"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
	"shanhu.io/misc/osutil"
	"shanhu.io/pisces/settings"
	"shanhu.io/virgo/dock"
)

type kernel struct {
	// For saving passwords and configs.
	settings *settings.Table

	// For registering domain routings.
	appDomains *appDomains

	// App registry.
	appRegistry *appRegistry

	// Applications.
	apps *apps
}

type drive struct {
	// Config file content.
	config *drvcfg.Config

	// Name of the endpoint (without leading '~').
	name string

	// User credential
	creds *creds.Endpoint

	// Remote server of homedrive.io, for downloading and credential management.
	server *url.URL

	// Uesr docker client
	dock *dock.Client

	// System docker client
	sysDock *dock.Client

	// HomeDrive kernel.
	*kernel

	// Mutex lock to block certain operations when one of the systems
	// operation is running.
	systemMu sync.Mutex
}

func parseServer(s string) (*url.URL, error) {
	if s == "" {
		return &url.URL{
			Scheme: "https",
			Host:   "www.homedrive.io",
		}, nil
	}
	return url.Parse(s)
}

func loadIdentity(f string) ([]byte, error) {
	if f == "" {
		f = "var/jarvis.pem"
	}
	return creds.ReadPrivateFile(f)
}

func newDrive(config *drvcfg.Config, k *kernel) (*drive, error) {
	key, err := loadIdentity(config.IdentityPem)
	if err != nil {
		return nil, errcode.Annotate(err, "load identity")
	}

	name := config.Name
	if name == "" {
		return nil, errcode.InvalidArgf("name is empty")
	}

	server, err := parseServer(config.Server)
	if err != nil {
		return nil, errcode.Annotate(err, "parse server URL")
	}

	userDockSock := config.DockerSock
	if userDockSock == "" {
		userDockSock = "/var/run/docker.sock"
	}
	sysDockSock := config.SystemDockerSock
	if sysDockSock == "" {
		sysDockSock = "/var/run/system-docker.sock"
	}
	hasSysDock, err := osutil.Exist(sysDockSock)
	if err != nil {
		return nil, errcode.Annotate(err, "check if system dock exists")
	}
	var sysDock *dock.Client
	if hasSysDock {
		sysDock = dock.NewUnixClient(sysDockSock)
	}

	ep := creds.NewRobot("~"+name, server.String(), "", nil)
	ep.Key = key

	return &drive{
		config:  config,
		name:    name,
		creds:   ep,
		server:  server,
		dock:    dock.NewUnixClient(userDockSock),
		sysDock: sysDock,
		kernel:  k,
	}, nil
}

func (d *drive) dialServer() (*httputil.Client, error) {
	return creds.DialEndpoint(d.creds)
}

func (d *drive) cont(s string) string {
	return drvcfg.Name(d.config.Naming, s)
}

func (d *drive) vol(s string) string {
	return drvcfg.Name(d.config.Naming, s)
}

func (d *drive) image(s string) string {
	return drvcfg.Image(d.config.Naming, s)
}

func (d *drive) network() string { return drvcfg.Network(d.config.Naming) }
func (d *drive) core() string    { return drvcfg.Core(d.config.Naming) }
func (d *drive) oldCore() string { return drvcfg.OldCore(d.config.Naming) }

func (d *drive) hasSys() bool { return d.sysDock != nil }

func (d *drive) burmilla() (*burmilla.Burmilla, error) {
	if d.sysDock == nil {
		return nil, errcode.Internalf("system-docker not found")
	}
	return burmilla.New(d.sysDock), nil
}

func (d *drive) appReflect(name string) (app, error) {
	stub, err := d.apps.stub(name)
	if err != nil {
		return nil, err
	}
	return stub.app, nil
}

func (d *drive) tags() string {
	var tags []string
	if !d.hasSys() {
		tags = append(tags, "soft")
	}
	if d.config.Naming == nil {
		tags = append(tags, "old-naming")
	}

	if len(tags) == 0 {
		return "-"
	}
	return strings.Join(tags, ",")
}
