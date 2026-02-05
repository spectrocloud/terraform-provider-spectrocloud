// Copyright (c) Spectro Cloud
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// Fixtures provides factory methods for creating test data
type Fixtures struct{}

// NewFixtures creates a new Fixtures instance
func NewFixtures() *Fixtures {
	return &Fixtures{}
}

// ProjectOption is a functional option for configuring a test project
type ProjectOption func(*models.V1ProjectEntity)

// Project creates a test project entity with optional configurations
func (f *Fixtures) Project(opts ...ProjectOption) *models.V1ProjectEntity {
	project := &models.V1ProjectEntity{
		Metadata: &models.V1ObjectMeta{
			Name:        "test-project",
			UID:         "",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}

	for _, opt := range opts {
		opt(project)
	}

	return project
}

// WithProjectName sets the project name
func WithProjectName(name string) ProjectOption {
	return func(p *models.V1ProjectEntity) {
		p.Metadata.Name = name
	}
}

// WithProjectUID sets the project UID
func WithProjectUID(uid string) ProjectOption {
	return func(p *models.V1ProjectEntity) {
		p.Metadata.UID = uid
	}
}

// WithProjectDescription sets the project description
func WithProjectDescription(description string) ProjectOption {
	return func(p *models.V1ProjectEntity) {
		if p.Metadata.Annotations == nil {
			p.Metadata.Annotations = make(map[string]string)
		}
		p.Metadata.Annotations["description"] = description
	}
}

// WithProjectLabels sets the project labels (tags)
func WithProjectLabels(labels map[string]string) ProjectOption {
	return func(p *models.V1ProjectEntity) {
		p.Metadata.Labels = labels
	}
}

// ProjectResponse creates a test project response (as returned by API)
func (f *Fixtures) ProjectResponse(opts ...ProjectResponseOption) *models.V1Project {
	project := &models.V1Project{
		Metadata: &models.V1ObjectMeta{
			Name:        "test-project",
			UID:         "test-uid-123",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Spec: &models.V1ProjectSpec{
			LogoURL: "",
		},
		Status: &models.V1ProjectStatus{
			IsDisabled: false,
		},
	}

	for _, opt := range opts {
		opt(project)
	}

	return project
}

// ProjectResponseOption is a functional option for configuring a test project response
type ProjectResponseOption func(*models.V1Project)

// WithResponseProjectName sets the project name in response
func WithResponseProjectName(name string) ProjectResponseOption {
	return func(p *models.V1Project) {
		p.Metadata.Name = name
	}
}

// WithResponseProjectUID sets the project UID in response
func WithResponseProjectUID(uid string) ProjectResponseOption {
	return func(p *models.V1Project) {
		p.Metadata.UID = uid
	}
}

// WithResponseProjectDescription sets the project description in response
func WithResponseProjectDescription(description string) ProjectResponseOption {
	return func(p *models.V1Project) {
		if p.Metadata.Annotations == nil {
			p.Metadata.Annotations = make(map[string]string)
		}
		p.Metadata.Annotations["description"] = description
	}
}

// WithResponseProjectLabels sets the project labels in response
func WithResponseProjectLabels(labels map[string]string) ProjectResponseOption {
	return func(p *models.V1Project) {
		p.Metadata.Labels = labels
	}
}
