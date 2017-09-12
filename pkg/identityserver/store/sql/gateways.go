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
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/utils"
)

// GatewayStore implements store.GatewayStore.
type GatewayStore struct {
	*Store
	factory factory.GatewayFactory
}

// ErrGatewayNotFound is returned when trying to fetch a gateway that does not exist.
var ErrGatewayNotFound = errors.New("gateway not found")

// ErrGatewayIDTaken is returned when trying to create a new gateway with an ID
// that already exists.
var ErrGatewayIDTaken = errors.New("gateway ID already taken")

// ErrGatewayAttributeNotFound is returned when trying to delete an attribute
// that does not exist.
var ErrGatewayAttributeNotFound = errors.New("gateway attribute not found")

// ErrGatewayAntennaNotFound is returned when trying to delete an antenna that
// does not exist.
var ErrGatewayAntennaNotFound = errors.New("gateway antenna not found")

// ErrGatewayCollaboratorNotFound is returned when trying to remove a
// collaborator that does not exist.
var ErrGatewayCollaboratorNotFound = errors.New("gateway collaborator not found")

// ErrGatewayCollaboratorRightNotFound is returned when trying to revoke a
// right from a collaborator that is not granted.
var ErrGatewayCollaboratorRightNotFound = errors.New("gateway collaborator right not found")

// Register creates a new Gateway and returns the new created Gateway.
func (s *GatewayStore) Register(gateway types.Gateway) (types.Gateway, error) {
	result := s.factory.Gateway()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.register(tx, gateway, result)
		if err != nil {
			return err
		}

		// store attributes
		gtw := gateway.GetGateway()
		for attribute, value := range gtw.Attributes {
			err := s.setAttribute(tx, gtw.ID, attribute, value)
			if err != nil {
				return err
			}
		}
		result.SetAttributes(gtw.Attributes)

		// store antennas
		for _, antenna := range gtw.Antennas {
			err := s.setAntenna(tx, gtw.ID, antenna)
			if err != nil {
				return err
			}
		}
		result.SetAntennas(gtw.Antennas)

		return nil
	})
	return result, err
}

func (s *GatewayStore) register(q db.QueryContext, gateway types.Gateway, result types.Gateway) error {
	gtw := gateway.GetGateway()
	err := q.NamedSelectOne(
		result,
		`INSERT
			INTO gateways (id, description, frequency_plan, key, activated,
					status_public, location_public, owner_public, auto_update,
					brand, model, routers)
			VALUES (:id, :description, :frequency_plan, :key, :activated,
					:status_public, :location_public, :owner_public, :auto_update,
					:brand, :model, :routers)
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

// FindByID finds a Gateway by ID and retrieves it.
func (s *GatewayStore) FindByID(gtwID string) (types.Gateway, error) {
	result := s.factory.Gateway()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.gateway(tx, gtwID, result)
		if err != nil {
			return err
		}

		attributes, err := s.listAttributes(tx, gtwID)
		if err != nil {
			return err
		}
		result.SetAttributes(attributes)

		antennas, err := s.listAntennas(tx, gtwID)
		if err != nil {
			return err
		}
		result.SetAntennas(antennas)

		return s.loadAttributes(tx, gtwID, result)
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

// FindByUser returns all the Gateways to which a given User is collaborator.
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

		attributes, err := s.listAttributes(q, gtwID)
		if err != nil {
			return err
		}
		gateway.SetAttributes(attributes)

		*result = append(*result, gateway)
	}
	return nil
}

// Edit updates the Gateway and returns the updated Gateway.
func (s *GatewayStore) Edit(gateway types.Gateway) (types.Gateway, error) {
	result := s.factory.Gateway()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.edit(tx, gateway, result)
		if err != nil {
			return err
		}

		gtw := gateway.GetGateway()

		for attribute, value := range gtw.Attributes {
			err := s.setAttribute(tx, gtw.ID, attribute, value)
			if err != nil {
				return err
			}
		}
		result.SetAttributes(gtw.Attributes)

		for _, antenna := range gtw.Antennas {
			err := s.setAntenna(tx, gtw.ID, antenna)
			if err != nil {
				return err
			}
		}
		result.SetAntennas(gtw.Antennas)

		return nil
	})
	return result, err
}

func (s *GatewayStore) edit(q db.QueryContext, gateway, result types.Gateway) error {
	gtw := gateway.GetGateway()
	err := q.NamedSelectOne(
		result,
		`UPDATE gateways
			SET description = :description, frequency_plan = :frequency_plan,
					key = :key, activated = :activated, status_public = :status_public,
					location_public = :location_public, owner_public = owner_public,
					auto_update = :auto_update, brand = :brand, model = :model,
					routers = :routers
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

// Archive disables a Gateway.
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

// SetAttribute inserts or modifies an existing Gateway attribute.
func (s *GatewayStore) SetAttribute(gtwID string, attribute, value string) error {
	return s.setAttribute(s.db, gtwID, attribute, value)
}

func (s *GatewayStore) setAttribute(q db.QueryContext, gtwID string, attribute, value string) error {
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

// ListAttributes returns all the Gateway attributes.
func (s *GatewayStore) ListAttributes(gtwID string) (map[string]string, error) {
	return s.listAttributes(s.db, gtwID)
}

func (s *GatewayStore) listAttributes(q db.QueryContext, gtwID string) (map[string]string, error) {
	var attrs []struct {
		Attribute string `db:"attribute"`
		Value     string `db:"value"`
	}
	err := q.Select(
		&attrs,
		`SELECT attribute, value
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

// RemoveAttribute removes a specific Gateway attribute.
func (s *GatewayStore) RemoveAttribute(gtwID string, attribute string) error {
	return s.removeAttribute(s.db, gtwID, attribute)
}

func (s *GatewayStore) removeAttribute(q db.QueryContext, gtwID string, attribute string) error {
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

// SetAntenna inserts or modifies an existing Gateway antenna.
func (s *GatewayStore) SetAntenna(gtwID string, antenna types.GatewayAntenna) error {
	return s.setAntenna(s.db, gtwID, antenna)
}

func (s *GatewayStore) setAntenna(q db.QueryContext, gtwID string, antenna types.GatewayAntenna) error {
	var ant struct {
		types.GatewayAntenna
		Longitude *float32
		Latitude  *float32
		Altitude  *int32
	}
	ant.GatewayAntenna = antenna

	if ant.GatewayAntenna.Location != nil {
		ant.Longitude = utils.Float32Address(ant.GatewayAntenna.Location.Longitude)
		ant.Latitude = utils.Float32Address(ant.GatewayAntenna.Location.Latitude)
		ant.Altitude = utils.Int32Address(ant.GatewayAntenna.Location.Altitude)
	}

	_, err := q.Exec(
		`INSERT
			INTO gateways_antennas (gateway_id, antenna_id, type, model,
					placement, longitude, latitude, altitude)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (gateway_id, antenna_id)
			DO UPDATE SET type = $3, model = $4, placement = $5, longitude = $6,
					latitude = $7, altitude = $8
			WHERE gateway_id = $1 AND antenna_id = $2`,
		gtwID,
		ant.ID,
		ant.Type,
		ant.Model,
		ant.Placement,
		ant.Longitude,
		ant.Latitude,
		ant.Altitude)
	return err
}

// ListAntennas returns all the registered antennas that belong to a certain Gateway.
func (s *GatewayStore) ListAntennas(gtwID string) ([]types.GatewayAntenna, error) {
	return s.listAntennas(s.db, gtwID)
}

func (s *GatewayStore) listAntennas(q db.QueryContext, gtwID string) ([]types.GatewayAntenna, error) {
	var antnns []struct {
		Longitude *float32 `db:"longitude"`
		Latitude  *float32 `db:"latitude"`
		Altitude  *int32   `db:"altitude"`
		types.GatewayAntenna
	}
	err := q.Select(
		&antnns,
		`SELECT antenna_id, longitude, latitude, altitude, type, model, placement
			FROM gateways_antennas
			WHERE gateway_id = $1`,
		gtwID)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	result := make([]types.GatewayAntenna, 0, len(antnns))
	for _, antenna := range antnns {
		result = append(result, types.GatewayAntenna{
			ID:        antenna.ID,
			Location:  helpers.Location(antenna.Longitude, antenna.Latitude, antenna.Altitude),
			Type:      antenna.Type,
			Model:     antenna.Model,
			Placement: antenna.Placement,
		})
	}

	return result, nil
}

// RemoveAntenna deletes an antenna from a gateway.
func (s *GatewayStore) RemoveAntenna(gtwID, antennaID string) error {
	return s.removeAntenna(s.db, gtwID, antennaID)
}

func (s *GatewayStore) removeAntenna(q db.QueryContext, gtwID, antennaID string) error {
	var i string
	err := q.SelectOne(
		&i,
		`DELETE
			FROM gateways_antennas
			WHERE gateway_id = $1 AND antenna_id = $2
			RETURNING antenna_id`,
		gtwID,
		antennaID)
	if db.IsNoRows(err) {
		return ErrGatewayAntennaNotFound
	}
	return err
}

// AddCollaborator adds a collaborator to a gateway.
func (s *GatewayStore) AddCollaborator(gtwID string, collaborator types.Collaborator) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.addCollaborator(tx, gtwID, collaborator)
	})
	return err
}

func (s *GatewayStore) addCollaborator(q db.QueryContext, gtwID string, collaborator types.Collaborator) error {
	for _, right := range collaborator.Rights {
		err := s.addRight(q, gtwID, collaborator.Username, right)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListCollaborators retrieves all the gateway collaborators.
func (s *GatewayStore) ListCollaborators(gtwID string) ([]types.Collaborator, error) {
	return s.listCollaborators(s.db, gtwID)
}

func (s *GatewayStore) listCollaborators(q db.QueryContext, gtwID string) ([]types.Collaborator, error) {
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

// ListOwners retrieves all the owners of a gateway.
func (s *GatewayStore) ListOwners(gtwID string) ([]string, error) {
	return s.listOwners(s.db, gtwID)
}

func (s *GatewayStore) listOwners(q db.QueryContext, gtwID string) ([]string, error) {
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

// RemoveCollaborator removes a collaborator from a gateway.
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

// AddRight grants a given right to a given User.
func (s *GatewayStore) AddRight(gtwID string, username string, right types.Right) error {
	return s.addRight(s.db, gtwID, username, right)
}

func (s *GatewayStore) addRight(q db.QueryContext, gtwID string, username string, right types.Right) error {
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

// ListUserRights returns the rights the User has for a gateway.
func (s *GatewayStore) ListUserRights(gtwID string, username string) ([]types.Right, error) {
	return s.listUserRights(s.db, gtwID, username)
}

func (s *GatewayStore) listUserRights(q db.QueryContext, gtwID string, username string) ([]types.Right, error) {
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

// RemoveRight revokes a given right from a given User.
func (s *GatewayStore) RemoveRight(gtwID string, username string, right types.Right) error {
	return s.removeRight(s.db, gtwID, username, right)
}

func (s *GatewayStore) removeRight(q db.QueryContext, gtwID string, username string, right types.Right) error {
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

// LoadAttributes loads the gateways attributes into result if it is an Attributer.
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

// WriteAttributes writes the gateways attributes into result if it is an Attributer.
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

// SetFactory allows to replace the DefaultGateway factory.
func (s *GatewayStore) SetFactory(factory factory.GatewayFactory) {
	s.factory = factory
}
