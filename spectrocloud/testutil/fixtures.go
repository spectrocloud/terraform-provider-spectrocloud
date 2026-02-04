// Copyright (c) Spectro Cloud
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// ProjectResponseOption is a functional option for building a V1Project response in tests.
type ProjectResponseOption func(*models.V1Project)

// WithResponseProjectName sets the project name in metadata.
func WithResponseProjectName(name string) ProjectResponseOption {
	return func(p *models.V1Project) {
		if p.Metadata == nil {
			p.Metadata = &models.V1ObjectMeta{}
		}
		p.Metadata.Name = name
	}
}

// WithResponseProjectUID sets the project UID in metadata.
func WithResponseProjectUID(uid string) ProjectResponseOption {
	return func(p *models.V1Project) {
		if p.Metadata == nil {
			p.Metadata = &models.V1ObjectMeta{}
		}
		p.Metadata.UID = uid
	}
}

// WithResponseProjectDescription sets the project description in metadata annotations.
func WithResponseProjectDescription(description string) ProjectResponseOption {
	return func(p *models.V1Project) {
		if p.Metadata == nil {
			p.Metadata = &models.V1ObjectMeta{}
		}
		if p.Metadata.Annotations == nil {
			p.Metadata.Annotations = make(map[string]string)
		}
		p.Metadata.Annotations["description"] = description
	}
}

// Fixtures provides factory methods for mock API responses in tests.
type Fixtures struct{}

// NewFixtures returns a new Fixtures instance.
func NewFixtures() *Fixtures {
	return &Fixtures{}
}

// ProjectResponse builds a *models.V1Project with the given options.
// Used by mock/unit tests that need a project response (e.g. GetProject).
func (f *Fixtures) ProjectResponse(opts ...ProjectResponseOption) *models.V1Project {
	p := &models.V1Project{
		Metadata: &models.V1ObjectMeta{},
		Spec:     &models.V1ProjectSpec{},
		Status:   &models.V1ProjectStatus{IsDisabled: false},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}
