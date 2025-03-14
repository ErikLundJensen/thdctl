/*
This file is similar to the API used by the Hetzner Dedicated Controller.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
)

// ServerParameters are the configurable fields of a server.
type ServerParameters struct {
	ServerNumber int    `json:"serverNumber"`
	Disk         string `json:"disk,omitempty"`
	TalosVersion string `json:"talosVersion,omitempty"`
	TalosImage   string `json:"talosImage,omitempty"`
}

// ServerObservation are the observable fields of a server.
type ServerObservation struct {
	ObservableField string `json:"observableField,omitempty"`
}

type TalosStatus struct {
	Status string `json:"status"`
}

// A ServerStatus represents the observed state of a server.
type ServerStatus struct {
	Details hetznerapi.ServerDetails `json:"details,omitempty"`
	Talos   TalosStatus              `json:"talos,omitempty"`
}

// A ServerSpec defines the desired state of a server.
type ServerSpec struct {
	ForProvider ServerParameters `json:"forProvider"`
}

type Server struct {
	Spec   ServerSpec   `json:"spec"`
	Status ServerStatus `json:"status,omitempty"`
}

// ServerList contains a list of server
type ServerList struct {
	Items []Server `json:"items"`
}
