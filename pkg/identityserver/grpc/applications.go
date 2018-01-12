// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

func (c *Client) CreateApplication(req *ttnpb.CreateApplicationRequest, credentials string) error {
	_, err := c.applications.CreateApplication(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) GetApplication(req *ttnpb.ApplicationIdentifier, credentials string) (*ttnpb.Application, error) {
	resp, err := c.applications.GetApplication(context.Background(), req, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListApplications(credentials string) (*ttnpb.ListApplicationsResponse, error) {
	resp, err := c.applications.ListApplications(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateApplication(req *ttnpb.UpdateApplicationRequest, credentials string) error {
	_, err := c.applications.UpdateApplication(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) DeleteApplication(req *ttnpb.ApplicationIdentifier, credentials string) error {
	_, err := c.applications.DeleteApplication(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) GenerateApplicationAPIKey(req *ttnpb.GenerateApplicationAPIKeyRequest, credentials string) (*ttnpb.APIKey, error) {
	resp, err := c.applications.GenerateApplicationAPIKey(context.Background(), req, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListApplicationAPIKeys(credentials string) (*ttnpb.ListApplicationAPIKeysResponse, error) {
	resp, err := c.applications.ListApplicationAPIKeys(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateApplicationAPIKey(req *ttnpb.UpdateApplicationAPIKeyRequest, credentials string) error {
	_, err := c.applications.UpdateApplicationAPIKey(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) RemoveApplicationAPIKey(req *ttnpb.RemoveApplicationAPIKeyRequest, credentials string) error {
	_, err := c.applications.RemoveApplicationAPIKey(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) SetApplicationCollaborator(req *ttnpb.ApplicationCollaborator, credentials string) error {
	_, err := c.applications.SetApplicationCollaborator(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) ListApplicationCollaborators(req *ttnpb.ApplicationIdentifier, credentials string) (*ttnpb.ListApplicationCollaboratorsResponse, error) {
	resp, err := c.applications.ListApplicationAPIKeys(context.Background(), req, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListApplicationRights(req *ttnpb.ApplicationIdentifier, credentials string) (*ttnpb.ListApplicationRightsResponse, error) {
	rights, err := c.cache.GetOrFetch(c.cache.ApplicationKey(req.ApplicationID), func() ([]ttnpb.Right, error) {
		resp, err := c.applications.ListApplicationRights(context.Background(), req, creds(c.md, credentials))
		if err != nil {
			return nil, err
		}

		return resp.Rights, nil
	})

	if err != nil {
		return nil, err
	}

	return &ttnpb.ListApplicationRightsResponse{
		Rights: rights,
	}, nil
}
