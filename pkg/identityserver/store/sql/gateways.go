// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/helpers"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// GatewayStore implements store.GatewayStore.
type GatewayStore struct {
	storer
}

func init() {
	ErrGatewayNotFound.Register()
	ErrGatewayIDTaken.Register()
}

// ErrGatewayNotFound is returned when trying to fetch a gateway that does not exist.
var ErrGatewayNotFound = &errors.ErrDescriptor{
	MessageFormat: "Gateway `{gateway_id}` does not exist",
	Code:          300,
	Type:          errors.NotFound,
	SafeAttributes: []string{
		"gateway_id",
	},
}

// ErrGatewayIDTaken is returned when trying to create a new gateway with an ID
// that already exists.
var ErrGatewayIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Gateway id `{gateway_id}` is already taken",
	Code:          301,
	Type:          errors.AlreadyExists,
	SafeAttributes: []string{
		"gateway_id",
	},
}

func NewGatewayStore(store storer) *GatewayStore {
	return &GatewayStore{
		storer: store,
	}
}

// Create creates a new gateway.
func (s *GatewayStore) Create(gateway types.Gateway) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.create(tx, gateway)
		if err != nil {
			return err
		}

		gtw := gateway.GetGateway()

		// store attributes
		err = s.setAttributes(tx, gtw.GatewayID, gtw.Attributes)
		if err != nil {
			return err
		}

		// store antennas
		err = s.addAntennas(tx, gtw.GatewayID, gtw.Antennas)
		if err != nil {
			return err
		}

		return s.writeAttributes(tx, gtw.GatewayID, gateway, nil)
	})
	return err
}

func (s *GatewayStore) create(q db.QueryContext, gateway types.Gateway) error {
	gtw := gateway.GetGateway()
	_, err := q.NamedExec(
		`INSERT
			INTO gateways (
					gateway_id,
					description,
					frequency_plan_id,
					activated_at,
					privacy_settings,
					auto_update,
					platform,
					cluster_address,
					archived_at)
			VALUES (
					:gateway_id,
					:description,
					:frequency_plan_id,
					:activated_at,
					:privacy_settings,
					:auto_update,
					:platform,
					:cluster_address,
					:archived_at)`,
		gtw)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrGatewayIDTaken.New(errors.Attributes{
			"gateway_id": gtw.GatewayID,
		})
	}

	return err
}

func (s *GatewayStore) addAntennas(q db.QueryContext, gtwID string, antennas []ttnpb.GatewayAntenna) error {
	if len(antennas) == 0 {
		return nil
	}
	query, args := s.addAntennasQuery(gtwID, antennas)
	_, err := q.Exec(query, args...)
	return err
}

func (s *GatewayStore) addAntennasQuery(gtwID string, antennas []ttnpb.GatewayAntenna) (string, []interface{}) {
	args := make([]interface{}, 1+7*len(antennas))
	args[0] = gtwID

	boundValues := make([]string, len(antennas))

	i := 0
	for j, antenna := range antennas {
		args[i+1] = antenna.Gain
		args[i+2] = antenna.Type
		args[i+3] = antenna.Model
		args[i+4] = antenna.Placement
		args[i+5] = antenna.Location.Longitude
		args[i+6] = antenna.Location.Latitude
		args[i+7] = antenna.Location.Altitude

		boundValues[j] = fmt.Sprintf("($1, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i+2, i+3, i+4, i+5, i+6, i+7, i+8)

		i += 7
	}

	query := fmt.Sprintf(
		`INSERT
			INTO gateways_antennas (
					gateway_id,
					gain,
					type,
					model,
					placement,
					longitude,
					latitude,
					altitude)
			VALUES %s`,
		strings.Join(boundValues, ", "))

	return query, args
}

// GetByID finds a gateway by ID and retrieves it.
func (s *GatewayStore) GetByID(gtwID string, factory store.GatewayFactory) (types.Gateway, error) {
	result := factory()

	err := s.transact(func(tx *db.Tx) error {
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

	if err != nil {
		return nil, err
	}

	return result, nil
}

// gateway fetchs a gateway from the database without antennas and attributes and
// saves it into result.
func (s *GatewayStore) gateway(q db.QueryContext, gtwID string, result types.Gateway) error {
	err := q.SelectOne(
		result,
		`SELECT *
				FROM gateways
				WHERE gateway_id = $1`,
		gtwID)
	if db.IsNoRows(err) {
		return ErrGatewayNotFound.New(errors.Attributes{
			"gateway_id": gtwID,
		})
	}
	return err
}

// FindByUser returns the Gateways to which an User is a collaborator.
func (s *GatewayStore) ListByUser(userID string, factory store.GatewayFactory) ([]types.Gateway, error) {
	var result []types.Gateway

	err := s.transact(func(tx *db.Tx) error {
		gateways, err := s.userGateways(tx, userID)
		if err != nil {
			return err
		}

		for _, gateway := range gateways {
			gtw := factory()
			*(gtw.GetGateway()) = gateway

			gtwID := gtw.GetGateway().GatewayID

			attributes, err := s.listAttributes(tx, gtwID)
			if err != nil {
				return err
			}
			gtw.SetAttributes(attributes)

			antennas, err := s.listAntennas(tx, gtwID)
			if err != nil {
				return err
			}
			gtw.SetAntennas(antennas)

			err = s.loadAttributes(tx, gtwID, gtw)
			if err != nil {
				return err
			}

			result = append(result, gtw)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *GatewayStore) userGateways(q db.QueryContext, userID string) ([]ttnpb.Gateway, error) {
	var gateways []ttnpb.Gateway
	err := q.Select(
		&gateways,
		`SELECT *
			FROM gateways
			WHERE gateway_id
			IN (
				SELECT
					DISTINCT gateway_id
					FROM gateways_collaborators
					WHERE user_id = $1
			)`,
		userID)

	if err != nil {
		return nil, err
	}

	if len(gateways) == 0 {
		return make([]ttnpb.Gateway, 0), nil
	}

	return gateways, nil
}

// Update updates the gateway.
func (s *GatewayStore) Update(gateway types.Gateway) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.update(tx, gateway)
		if err != nil {
			return err
		}

		gtw := gateway.GetGateway()

		err = s.updateAttributes(tx, gtw.GatewayID, gtw.Attributes)
		if err != nil {
			return err
		}

		err = s.updateAntennas(tx, gtw.GatewayID, gtw.Antennas)
		if err != nil {
			return err
		}

		return s.writeAttributes(tx, gtw.GatewayID, gateway, nil)
	})
	return err
}

func (s *GatewayStore) update(q db.QueryContext, gateway types.Gateway) error {
	gtw := gateway.GetGateway()

	var id string
	err := q.NamedSelectOne(
		&id,
		`UPDATE gateways
			SET description = :description,
				frequency_plan_id = :frequency_plan_id,
				activated_at = :activated_at,
				privacy_settings = :privacy_settings,
				auto_update = :auto_update,
				platform = :platform,
				cluster_address = :cluster_address,
				updated_at = current_timestamp()
			WHERE gateway_id = :gateway_id
			RETURNING gateway_id`,
		gtw)

	if db.IsNoRows(err) {
		return ErrGatewayNotFound.New(errors.Attributes{
			"gateway_id": gtw.GatewayID,
		})
	}

	return err
}

func (s *GatewayStore) updateAntennas(q db.QueryContext, gtwID string, antennas []ttnpb.GatewayAntenna) error {
	_, err := q.Exec("DELETE FROM gateways_antennas WHERE gateway_id = $1", gtwID)
	if err != nil {
		return err
	}

	return s.addAntennas(q, gtwID, antennas)
}

// Archive disables a Gateway.
func (s *GatewayStore) Archive(gtwID string) error {
	return s.archive(s.queryer(), gtwID)
}

func (s *GatewayStore) archive(q db.QueryContext, gtwID string) error {
	var id string
	err := q.SelectOne(
		&id,
		`UPDATE gateways
			SET archived_at = current_timestamp()
			WHERE gateway_id = $1
			returning gateway_id`,
		gtwID)
	if db.IsNoRows(err) {
		return ErrGatewayNotFound.New(errors.Attributes{
			"gateway_id": gtwID,
		})
	}
	return err
}

// updateAttributes removes the attributes that no longer exists for the gateway
// given its ID and sets the rest of attributes.
func (s *GatewayStore) updateAttributes(q db.QueryContext, gtwID string, attributes map[string]string) error {
	query, args := s.removeAttributeDiffQuery(gtwID, attributes)
	_, err := q.Exec(query, args...)
	if err != nil {
		return err
	}

	return s.setAttributes(q, gtwID, attributes)
}

// removeAttributeDiffQuery is the helper that construct the query to remove
// those gateway attributes that no longer exists. It returns the query together
// with the arguments list.
func (s *GatewayStore) removeAttributeDiffQuery(gtwID string, attributes map[string]string) (string, []interface{}) {
	args := make([]interface{}, 1+len(attributes))
	args[0] = gtwID

	boundVariables := make([]string, len(attributes))

	i := 0
	for k := range attributes {
		args[i+1] = k
		boundVariables[i] = fmt.Sprintf("$%d", i+2)
		i++
	}

	query := fmt.Sprintf(
		`DELETE
			FROM gateways_attributes
			WHERE gateway_id = $1 AND attribute NOT IN (%s)`,
		strings.Join(boundVariables, ", "))

	return query, args
}

// setAttributes inserts or modifies the attributes.
func (s *GatewayStore) setAttributes(q db.QueryContext, gtwID string, attributes map[string]string) error {
	query, args := s.setAttributesQuery(gtwID, attributes)
	_, err := q.Exec(query, args...)
	return err
}

// setAttributesQuery is a helper that constructs the upsert query for the
// setAttributes method and returns it together with the list of arguments.
func (s *GatewayStore) setAttributesQuery(gtwID string, attributes map[string]string) (string, []interface{}) {
	args := make([]interface{}, 1+2*len(attributes))
	args[0] = gtwID

	boundValues := make([]string, len(attributes))

	i := 1
	j := 0
	for k, v := range attributes {
		args[i] = k
		args[i+1] = v
		boundValues[j] = fmt.Sprintf("($1, $%d, $%d)", i+1, i+2)

		i += 2
		j += 1
	}

	query := fmt.Sprintf(
		`UPSERT
			INTO gateways_attributes (gateway_id, attribute, value)
			VALUES %s`,
		strings.Join(boundValues, ", "))

	return query, args
}

func (s *GatewayStore) listAttributes(q db.QueryContext, gtwID string) (map[string]string, error) {
	var attrs []struct {
		Attribute string
		Value     string
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

func (s *GatewayStore) listAntennas(q db.QueryContext, gtwID string) ([]ttnpb.GatewayAntenna, error) {
	var antnns []struct {
		Longitude float32
		Latitude  float32
		Altitude  int32
		ttnpb.GatewayAntenna
	}
	err := q.Select(
		&antnns,
		`SELECT longitude, latitude, altitude, gain, type, model, placement
			FROM gateways_antennas
			WHERE gateway_id = $1
			ORDER BY created_at ASC`,
		gtwID)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	result := make([]ttnpb.GatewayAntenna, 0, len(antnns))
	for _, antenna := range antnns {
		result = append(result, ttnpb.GatewayAntenna{
			Location: ttnpb.Location{
				Longitude: antenna.Longitude,
				Latitude:  antenna.Latitude,
				Altitude:  antenna.Altitude,
			},
			Gain:      antenna.GatewayAntenna.Gain,
			Type:      antenna.GatewayAntenna.Type,
			Model:     antenna.GatewayAntenna.Model,
			Placement: antenna.GatewayAntenna.Placement,
		})
	}

	return result, nil
}

// SetCollaborator inserts or modifies a collaborator within an entity.
// If the provided list of rights is empty the collaborator will be unset.
func (s *GatewayStore) SetCollaborator(collaborator ttnpb.GatewayCollaborator) error {
	if len(collaborator.Rights) == 0 {
		return s.unsetCollaborator(s.queryer(), collaborator.GatewayID, collaborator.UserID)
	}

	err := s.transact(func(tx *db.Tx) error {
		return s.setCollaborator(tx, collaborator)
	})
	return err
}

func (s *GatewayStore) unsetCollaborator(q db.QueryContext, gtwID, userID string) error {
	_, err := q.Exec(
		`DELETE
			FROM gateways_collaborators
			WHERE gateway_id = $1 AND user_id = $2`, gtwID, userID)
	return err
}

func (s *GatewayStore) setCollaborator(q db.QueryContext, collaborator ttnpb.GatewayCollaborator) error {
	query, args := s.removeRightsDiffQuery(collaborator)
	_, err := q.Exec(query, args...)
	if err != nil {
		return err
	}

	query, args = s.addRightsQuery(collaborator.GatewayID, collaborator.UserID, collaborator.Rights)
	_, err = q.Exec(query, args...)

	return err
}

func (s *GatewayStore) removeRightsDiffQuery(collaborator ttnpb.GatewayCollaborator) (string, []interface{}) {
	args := make([]interface{}, 2+len(collaborator.Rights))
	args[0] = collaborator.GatewayID
	args[1] = collaborator.UserID

	boundVariables := make([]string, len(collaborator.Rights))

	for i, right := range collaborator.Rights {
		args[i+2] = right
		boundVariables[i] = fmt.Sprintf("$%d", i+3)
	}

	query := fmt.Sprintf(
		`DELETE
			FROM gateways_collaborators
			WHERE gateway_id = $1 AND user_id = $2 AND "right" NOT IN (%s)`,
		strings.Join(boundVariables, ", "))

	return query, args
}

func (s *GatewayStore) addRightsQuery(gtwID, userID string, rights []ttnpb.Right) (string, []interface{}) {
	args := make([]interface{}, 2+len(rights))
	args[0] = gtwID
	args[1] = userID

	boundValues := make([]string, len(rights))

	for i, right := range rights {
		args[i+2] = right
		boundValues[i] = fmt.Sprintf("($1, $2, $%d)", i+3)
	}

	query := fmt.Sprintf(
		`INSERT
			INTO gateways_collaborators (gateway_id, user_id, "right")
			VALUES %s
			ON CONFLICT (gateway_id, user_id, "right")
			DO NOTHING`,
		strings.Join(boundValues, " ,"))

	return query, args
}

// ListCollaborators retrieves all the collaborators from an entity.
func (s *GatewayStore) ListCollaborators(gtwID string) ([]ttnpb.GatewayCollaborator, error) {
	return s.listCollaborators(s.queryer(), gtwID)
}

func (s *GatewayStore) listCollaborators(q db.QueryContext, gtwID string) ([]ttnpb.GatewayCollaborator, error) {
	var collaborators []struct {
		ttnpb.GatewayCollaborator
		Right ttnpb.Right
	}
	err := q.Select(
		&collaborators,
		`SELECT user_id, "right"
			FROM gateways_collaborators
			WHERE gateway_id = $1`,
		gtwID)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	byUser := make(map[string]*ttnpb.GatewayCollaborator)
	for _, collaborator := range collaborators {
		if _, exists := byUser[collaborator.UserID]; !exists {
			byUser[collaborator.UserID] = &ttnpb.GatewayCollaborator{
				GatewayIdentifier: ttnpb.GatewayIdentifier{gtwID},
				UserIdentifier:    ttnpb.UserIdentifier{collaborator.UserID},
				Rights:            []ttnpb.Right{collaborator.Right},
			}
			continue
		}

		byUser[collaborator.UserID].Rights = append(byUser[collaborator.UserID].Rights, collaborator.Right)
	}

	result := make([]ttnpb.GatewayCollaborator, 0, len(byUser))
	for _, collaborator := range byUser {
		result = append(result, *collaborator)
	}

	return result, nil
}

// ListUserRights returns the rights a given user has for an entity.
func (s *GatewayStore) ListUserRights(gtwID string, userID string) ([]ttnpb.Right, error) {
	return s.listUserRights(s.queryer(), gtwID, userID)
}

func (s *GatewayStore) listUserRights(q db.QueryContext, gtwID string, userID string) ([]ttnpb.Right, error) {
	var rights []ttnpb.Right
	err := q.Select(
		&rights,
		`SELECT "right"
			FROM gateways_collaborators
			WHERE gateway_id = $1 AND user_id = $2`,
		gtwID,
		userID)

	return rights, err
}

// LoadAttributes loads the gateways attributes into result if it is an Attributer.
func (s *GatewayStore) LoadAttributes(gateway types.Gateway) error {
	return s.loadAttributes(s.queryer(), gateway.GetGateway().GatewayID, gateway)
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
	return s.writeAttributes(s.queryer(), gateway.GetGateway().GatewayID, gateway, result)
}

func (s *GatewayStore) writeAttributes(q db.QueryContext, gtwID string, gateway types.Gateway, result types.Gateway) error {
	attr, ok := gateway.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		query, values := helpers.WriteAttributes(attr, namespace, "gateways", "gateway_id", gateway.GetGateway().GatewayID)

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
