// Copyright (C) 2023  Shanhu Tech Inc.
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

package drvconfig

// SystemInfo contains the high-level information of the board and the base
// OS image.
type SystemInfo struct {
	Board string `json:",omitempty"`
}

// Common boards.
const (
	BoardRpi4         = "rpi4"
	BoardNUC7         = "nuc7"
	BoardNUC10        = "nuc10"
	BoardDigitalOcean = "docn"
)
