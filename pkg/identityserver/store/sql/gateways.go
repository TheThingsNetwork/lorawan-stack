// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"errors"
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/helpers"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// GatewayStore implements store.GatewayStore
type GatewayStore struct {
	*Store
	factory factory.GatewayFactory
}

// ErrGatewayNotFound is returned when trying to fetch a gateway that does not exist
var ErrGatewayNotFound = errors.New("gateway not found")

// ErrGatewayIDTaken is returned when trying to create a new gateway with an ID
// that already exists
var ErrGatewayIDTaken = errors.New("gateway ID already taken")

// ErrGatewayAttributeNotFound is returned when trying to delete an attribute
// that does not exist
var ErrGatewayAttributeNotFound = errors.New("gateway attribute not found")

// ErrGatewayCollaboratorNotFound is returned when trying to remove a
// collaborator that does not exist
var ErrGatewayCollaboratorNotFound = errors.New("gateway collaborator not found")

// ErrGatewayCollaboratorRightNotFound is returned when trying to revoke a
// right from a collaborator that is not granted
var ErrGatewayCollaboratorRightNotFound = errors.New("gateway collaborator right not found")

// SetFactory replaces the factory
func (s *GatewayStore) SetFactory(factory factory.GatewayFactory) {
	s.factory = factory
}

// LoadAttributes loads the gateways attributes into result if it is an
// Attributer
func (s *GatewayStore) LoadAttributes(gateway types.Gateway) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.loadAttributes(tx, gateway.GetGateway().ID, gateway)
	})
}

func (s *GatewayStore) loadAttributes(q db.QueryContext, gtwID string, gateway types.Gateway) error {
	attr, ok := gateway.(store.Attributer)
	if !ok {
		return nil
	}

	// fill the gateway from all specified namespaces
	for _, namespace := range attr.Namespaces() {
		m := make(map[string]interface{})
		err := q.SelectOne(
			&m,
			fmt.Sprintf("SELECT * FROM %s_gateways WHERE gateway_id = $1", namespace),
			gtwID)
		if err != nil {
			return err
		}

		err = attr.Fill(namespace, m)
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteAttributes writes the gateways attributes into result if it is an gatewayAttributer
func (s *GatewayStore) WriteAttributes(gateway types.Gateway, result types.Gateway) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.writeAttributes(tx, gateway.GetGateway().ID, gateway, result)
	})
}

func (s *GatewayStore) writeAttributes(q db.QueryContext, gtwID string, gateway types.Gateway, result types.Gateway) error {
	attr, ok := gateway.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		query, values := helpers.WriteAttributes(attr, namespace, "gateways", "gateway_id", gateway.GetGateway().ID)

		r := make(map[string]interface{})
		err := q.SelectOne(r, query, values...)
		if err != nil {
			return err
		}

		if rattr, ok := result.(store.Attributer); ok {
			err = rattr.Fill(namespace, r)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// FindByID finds the gateway by ID
func (s *GatewayStore) FindByID(gtwID string) (types.Gateway, error) {
	result := s.factory.Gateway()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.gateway(tx, gtwID, result)
		if err != nil {
			return err
		}

		attributes, err := s.attributes(tx, gtwID)
		if err != nil {
			return err
		}
		result.SetAttributes(attributes)

		return nil
	})
	return result, err
}

func (s *GatewayStore) gateway(q db.QueryContext, gtwID string, result types.Gateway) error {
	err := q.SelectOne(result, "SELECT * FROM gateways WHERE id = $1", gtwID)
	if db.IsNoRows(err) {
		return ErrGatewayNotFound
	}
	return err
}

// Attributes returns the gateway attributes in a map
func (s *GatewayStore) Attributes(q db.QueryContext, gtwID string) (map[string]string, error) {
	return s.attributes(s.db, gtwID)
}

func (s *GatewayStore) attributes(q db.QueryContext, gtwID string) (map[string]string, error) {
	var attrs []struct {
		Attribute string `db:"attribute"`
		Value     string `db:"value"`
	}
	err := q.Select(
		attrs,
		`SELECT *
			FROM gateways_attributes
			WHERE gateway_id = $1`,
		gtwID)

	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, attr := range attrs {
		result[attr.Attribute] = attr.Value
	}

	return result, nil
}

// UpsertAttribute inserts or modifies an existing attribute
func (s *GatewayStore) UpsertAttribute(gtwID string, attribute, value string) error {
	return s.upsertAttribute(s.db, gtwID, attribute, value)
}

func (s *GatewayStore) upsertAttribute(q db.QueryContext, gtwID string, attribute, value string) error {
	_, err := q.Exec(
		`INSERT
			INTO gateways_attributes (gateway_id, attribute, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (gateway_id, attribute)
			DO UPDATE SET value = $3
			WHERE gateway_id = $1 AND attribute = $2`,
		gtwID,
		attribute,
		value)
	return err
}

// DeleteAttribute deletes an attribute
func (s *GatewayStore) DeleteAttribute(gtwID string, attribute string) error {
	return s.deleteAttribute(s.db, gtwID, attribute)
}

func (s *GatewayStore) deleteAttribute(q db.QueryContext, gtwID string, attribute string) error {
	var i string
	err := q.SelectOne(
		&i,
		`DELETE
			FROM gateways_attributes
			WHERE gateway_id = $1 AND attribute = $2
			RETURNING gateway_id`,
		gtwID,
		attribute)
	if db.IsNoRows(err) {
		return ErrGatewayAttributeNotFound
	}
	return err
}

// FindByUser returns the gateways to which an user is a collaborator
func (s *GatewayStore) FindByUser(username string) ([]types.Gateway, error) {
	var gateways []types.Gateway
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.userGateways(tx, username, &gateways)
	})
	return gateways, err
}

func (s *GatewayStore) userGateways(q db.QueryContext, username string, result *[]types.Gateway) error {
	var gtwIDs []string
	err := q.Select(
		&gtwIDs,
		`SELECT DISTINCT gateway_id
			FROM gateways_collaborators
			WHERE username = $1`,
		username)

	if !db.IsNoRows(err) && err != nil {
		return err
	}

	for _, gtwID := range gtwIDs {
		gateway := s.factory.Gateway()
		err := s.gateway(q, gtwID, gateway)
		if err != nil {
			return err
		}

		attributes, err := s.attributes(q, gtwID)
		if err != nil {
			return err
		}
		gateway.SetAttributes(attributes)

		*result = append(*result, gateway)
	}
	return nil
}

// Create creates a new gateway and returns the resulting gateway
func (s *GatewayStore) Create(gateway types.Gateway) (types.Gateway, error) {
	result := s.factory.Gateway()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.create(tx, gateway, result)
		if err != nil {
			return err
		}

		// store attributes
		gtw := gateway.GetGateway()
		for attribute, value := range gtw.Attributes {
			err := s.upsertAttribute(tx, gtw.ID, attribute, value)
			if err != nil {
				return err
			}
		}
		result.SetAttributes(gtw.Attributes)

		return nil
	})
	return result, err
}

func (s *GatewayStore) create(q db.QueryContext, gateway types.Gateway, result types.Gateway) error {
	gtw := gateway.GetGateway()
	err := q.NamedSelectOne(
		result,
		`INSERT
			INTO gateways (id, description, frequency_plan, key, activated,
					status_public, location_public, owner_public, auto_update,
					brand, model, antenna_type, antenna_model, antenna_placement,
					antenna_altitude, antenna_location, routers)
			VALUES (:id, :description, :frequency_plan, :key, :activated,
					:status_public, :location_public, :owner_public, :auto_update,
					:brand, :model, :antenna_type, :antenna_model, :antenna_placement,
					:antenna_altitude, :antenna_location, :routers)
			RETURNING *`,
		gtw)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrGatewayIDTaken
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, gtw.ID, gateway, result)
}

// Update updates a gateway and returns the updated version
func (s *GatewayStore) Update(gateway types.Gateway) (types.Gateway, error) {
	result := s.factory.Gateway()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.update(tx, gateway, result)
		if err != nil {
			return err
		}

		for attribute, value := range gateway.GetGateway().Attributes {
			err := s.upsertAttribute(tx, gateway.GetGateway().ID, attribute, value)
			if err != nil {
				return err
			}
		}
		result.SetAttributes(gateway.GetGateway().Attributes)

		return nil
	})
	return result, err
}

func (s *GatewayStore) update(q db.QueryContext, gateway, result types.Gateway) error {
	gtw := gateway.GetGateway()
	err := q.SelectOne(
		result,
		`UPDATE gateways
			SET description = :description, frequency_plan = :frequency_plan, key = :key,
					activated = :activated, status_public = :status_public, location_public = :location_public,
					owner_public = owner_public, auto_update = :auto_update, brand = :brand, model = :model,
					antenna_type = :antenna_type, antenna_model = :antenna_model, antenna_placement = :antenna_placement,
					antenna_altitude = :antenna_altitude, antenna_location = :antenna_location, routers = :routers
			WHERE id = :id
			RETURNING *`,
		gtw)

	if db.IsNoRows(err) {
		return ErrGatewayNotFound
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, gtw.ID, gateway, result)
}

// Archive archives a gateway
func (s *GatewayStore) Archive(gtwID string) error {
	return s.archive(s.db, gtwID)
}

func (s *GatewayStore) archive(q db.QueryContext, gtwID string) error {
	var id string
	err := q.SelectOne(
		&id,
		`UPDATE gateways
			SET archived = $1
			WHERE id = $2
			RETURNING id`,
		time.Now(),
		gtwID)
	if db.IsNoRows(err) {
		return ErrGatewayNotFound
	}
	return err
}

// Owners returns a list of users who have owners rights to a given gateway
func (s *GatewayStore) Owners(gtwID string) ([]string, error) {
	return s.owners(s.db, gtwID)
}

func (s *GatewayStore) owners(q db.QueryContext, gtwID string) ([]string, error) {
	var owners []string
	err := q.Select(
		&owners,
		`SELECT username
			FROM gateways_collaborators
			WHERE gateway_id = $1 AND "right" = $2`,
		gtwID,
		types.GatewayOwnerRight)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}
	return owners, nil
}

// Collaborators returns the list of collaborators to a given gateway
func (s *GatewayStore) Collaborators(gtwID string) ([]types.Collaborator, error) {
	return s.collaborators(s.db, gtwID)
}

func (s *GatewayStore) collaborators(q db.QueryContext, gtwID string) ([]types.Collaborator, error) {
	var collaborators []struct {
		types.Collaborator
		Right string `db:"right"`
	}
	err := q.Select(
		&collaborators,
		`SELECT username, "right"
			FROM gateways_collaborators
			WHERE gateway_id = $1`,
		gtwID)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	byUser := make(map[string]*types.Collaborator)
	for _, collaborator := range collaborators {
		if _, exists := byUser[collaborator.Username]; !exists {
			byUser[collaborator.Username] = &types.Collaborator{
				Username: collaborator.Username,
				Rights:   []types.Right{types.Right(collaborator.Right)},
			}
			continue
		}

		byUser[collaborator.Username].Rights = append(byUser[collaborator.Username].Rights, types.Right(collaborator.Right))
	}

	result := make([]types.Collaborator, 0, len(byUser))
	for _, collaborator := range byUser {
		result = append(result, *collaborator)
	}

	return result, nil
}

// AddCollaborator adds a new collaborator to a given gateway
func (s *GatewayStore) AddCollaborator(gtwID string, collaborator types.Collaborator) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.addCollaborator(tx, gtwID, collaborator)
	})
	return err
}

func (s *GatewayStore) addCollaborator(q db.QueryContext, gtwID string, collaborator types.Collaborator) error {
	for _, right := range collaborator.Rights {
		err := s.grantRight(q, gtwID, collaborator.Username, right)
		if err != nil {
			return err
		}
	}
	return nil
}

// GrantRight grants a right to a specific user in a given gateway
func (s *GatewayStore) GrantRight(gtwID string, username string, right types.Right) error {
	return s.grantRight(s.db, gtwID, username, right)
}

func (s *GatewayStore) grantRight(q db.QueryContext, gtwID string, username string, right types.Right) error {
	_, err := q.Exec(
		`INSERT
			INTO gateways_collaborators (gateway_id, username, "right")
			VALUES ($1, $2, $3)
			ON CONFLICT (gateway_id, username, "right")
			DO NOTHING`,
		gtwID,
		username,
		right)
	return err
}

// RevokeRight revokes a specific right to a specific user in a given gateway
func (s *GatewayStore) RevokeRight(gtwID string, username string, right types.Right) error {
	return s.revokeRight(s.db, gtwID, username, right)
}

func (s *GatewayStore) revokeRight(q db.QueryContext, gtwID string, username string, right types.Right) error {
	var u string
	err := q.SelectOne(
		&u,
		`DELETE
			FROM gateways_collaborators
			WHERE gateway_id = $1 AND username = $2 AND "right" = $3
			RETURNING username`,
		gtwID,
		username,
		right)
	if db.IsNoRows(err) {
		return ErrGatewayCollaboratorRightNotFound
	}
	return err
}

// RemoveCollaborator removes a collaborator of a given gateway
func (s *GatewayStore) RemoveCollaborator(gtwID string, username string) error {
	return s.removeCollaborator(s.db, gtwID, username)
}

func (s *GatewayStore) removeCollaborator(q db.QueryContext, gtwID string, username string) error {
	var u string
	err := q.SelectOne(
		&u,
		`DELETE
			FROM gateways_collaborators
			WHERE gateway_id = $1 AND username = $2
			RETURNING username`,
		gtwID,
		username)
	if db.IsNoRows(err) {
		return ErrGatewayCollaboratorNotFound
	}
	return err
}

// UserRights returns the list of rights that an user has to a given gateway
func (s *GatewayStore) UserRights(gtwID string, username string) ([]types.Right, error) {
	return s.userRights(s.db, gtwID, username)
}

func (s *GatewayStore) userRights(q db.QueryContext, gtwID string, username string) ([]types.Right, error) {
	var rights []types.Right
	err := q.Select(
		&rights,
		`SELECT "right"
			FROM gateways_collaborators
			WHERE gateway_id = $1 AND username = $2`,
		gtwID,
		username)
	return rights, err
}
