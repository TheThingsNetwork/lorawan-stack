// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

func (c *Client) CreateUser(req *ttnpb.CreateUserRequest) error {
	_, err := c.users.CreateUser(context.Background(), req)
	return err
}

func (c *Client) GetUser(credentials string) (*ttnpb.User, error) {
	resp, err := c.users.GetUser(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateUser(req *ttnpb.UpdateUserRequest, credentials string) error {
	_, err := c.users.UpdateUser(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) UpdateUserPasssword(req *ttnpb.UpdateUserPasswordRequest, credentials string) error {
	_, err := c.users.UpdateUserPassword(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) DeleteUser(credentials string) error {
	_, err := c.users.DeleteUser(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	return err
}

func (c *Client) DeleteUser(credentials string) error {
	_, err := c.users.DeleteUser(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	return err
}

func (c *Client) GenerateUserAPIKey(req *ttnpb.GenerateUserAPIKeyRequest, credentials string) (*ttnpb.APIKey, error) {
	resp, err := c.users.GenerateUserAPIKey(context.Background(), req, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListUserAPIKeys(credentials string) (*ttnpb.ListUserAPIKeysResponse, error) {
	resp, err := c.users.ListUserAPIKeys(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateUserAPIKey(req *ttnpb.UpdateUserAPIKeyRequest, credentials string) error {
	_, err := c.users.UpdateUserAPIKey(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) RemoveUserAPIKey(req *ttnpb.RemoveUserAPIKeyRequest, credentials string) error {
	_, err := c.users.RemoveUserAPIKey(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) ValidateUserEmail(req *ttnpb.ValidateUserEmailRequest) error {
	_, err := c.users.ValidateUserEmail(context.Background(), req)
	return err
}

func (c *Client) RequestUserEmailValidation(credentials string) error {
	_, err := c.users.RequestUserEmailValidation(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	return err
}

func (c *Client) ListAuthorizedClients(credentials string) (*ttnpb.ListAuthorizedClientsResponse, error) {
	resp, err := c.users.ListAuthorizedClients(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) RevokeAuthorizedClient(req *ttnpb.ClientIdentifier, credentials string) error {
	_, err := c.users.RevokeAuthorizedClient(context.Background(), req, creds(c.md, credentials))
	return err
}
