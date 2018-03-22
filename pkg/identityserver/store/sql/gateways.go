// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

type gateway struct {
	ID uuid.UUID
	ttnpb.Gateway
}

// GatewayStore implements store.GatewayStore.
type GatewayStore struct {
	storer
	*extraAttributesStore
	*apiKeysStore

	*accountStore
}

// NewGatewayStore returns a GatewayStore.
func NewGatewayStore(store storer) *GatewayStore {
	return &GatewayStore{
		storer:               store,
		extraAttributesStore: newExtraAttributesStore(store, "gateway"),
		apiKeysStore:         newAPIKeysStore(store, "gateway"),
		accountStore:         newAccountStore(store),
	}
}

func (s *GatewayStore) getGatewayIdentifiersFromID(q db.QueryContext, id uuid.UUID) (res ttnpb.GatewayIdentifiers, err error) {
	err = q.SelectOne(
		&res,
		`SELECT
				gateway_id
			FROM gateways
			WHERE id = $1`,
		id)
	return
}

// getGatewayID returns the UUID of the gateway that matches the identifier.
func (s *GatewayStore) getGatewayID(q db.QueryContext, ids ttnpb.GatewayIdentifiers) (res uuid.UUID, err error) {
	err = q.SelectOne(
		&res,
		`SELECT
				id
			FROM gateways
			WHERE gateway_id = $1`,
		ids.GatewayID)
	if db.IsNoRows(err) {
		err = ErrGatewayNotFound.New(nil)
	}
	return
}

// Create creates a new gateway.
func (s *GatewayStore) Create(gateway store.Gateway) error {
	err := s.transact(func(tx *db.Tx) error {
		gtw := gateway.GetGateway()

		gtwID, err := s.create(tx, gtw)
		if err != nil {
			return err
		}

		// store attributes
		err = s.setAttributes(tx, gtwID, gtw.Attributes)
		if err != nil {
			return err
		}

		// store antennas
		err = s.addAntennas(tx, gtwID, gtw.Antennas)
		if err != nil {
			return err
		}

		// store radios
		err = s.addRadios(tx, gtwID, gtw.Radios)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, gtwID, gateway)
	})
	return err
}

func (s *GatewayStore) create(q db.QueryContext, gateway *ttnpb.Gateway) (id uuid.UUID, err error) {
	err = q.NamedSelectOne(
		&id,
		`INSERT
			INTO gateways (
					gateway_id,
					description,
					frequency_plan_id,
					activated_at,
					privacy_settings,
					auto_update,
					platform,
					cluster_address)
			VALUES (
					:gateway_id,
					:description,
					:frequency_plan_id,
					:activated_at,
					:privacy_settings,
					:auto_update,
					:platform,
					:cluster_address)
			RETURNING id`,
		gateway)
	if _, yes := db.IsDuplicate(err); yes {
		err = ErrGatewayIDTaken.New(nil)
	}
	return
}

func (s *GatewayStore) addAntennas(q db.QueryContext, gtwID uuid.UUID, antennas []ttnpb.GatewayAntenna) error {
	if len(antennas) == 0 {
		return nil
	}
	query, args := s.addAntennasQuery(gtwID, antennas)
	_, err := q.Exec(query, args...)
	return err
}

func (s *GatewayStore) addAntennasQuery(gtwID uuid.UUID, antennas []ttnpb.GatewayAntenna) (string, []interface{}) {
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

func (s *GatewayStore) addRadios(q db.QueryContext, gtwID uuid.UUID, radios []ttnpb.GatewayRadio) error {
	if len(radios) == 0 {
		return nil
	}
	query, args := s.addRadiosQuery(gtwID, radios)
	_, err := q.Exec(query, args...)
	return err
}

func (s *GatewayStore) addRadiosQuery(gtwID uuid.UUID, radios []ttnpb.GatewayRadio) (string, []interface{}) {
	args := make([]interface{}, 1+2*len(radios))
	args[0] = gtwID

	boundValues := make([]string, len(radios))

	i := 0
	for j, radio := range radios {
		args[i+1] = radio.Frequency
		args[i+2] = radio.TxConfiguration

		boundValues[j] = fmt.Sprintf("($1, $%d, $%d)", i+2, i+3)

		i += 2
	}

	query := fmt.Sprintf(
		`INSERT
			INTO gateways_radios (
					gateway_id,
					frequency,
					tx_configuration)
			VALUES %s`,
		strings.Join(boundValues, ", "))

	return query, args
}

// GetByID finds a gateway by ID and retrieves it.
func (s *GatewayStore) GetByID(ids ttnpb.GatewayIdentifiers, specializer store.GatewaySpecializer) (result store.Gateway, err error) {
	err = s.transact(func(tx *db.Tx) error {
		gtwID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		gateway, err := s.getByID(tx, gtwID)
		if err != nil {
			return err
		}

		result = specializer(*gateway)

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

		radios, err := s.listRadios(tx, gtwID)
		if err != nil {
			return err
		}
		result.SetRadios(radios)

		return s.loadAttributes(tx, gtwID, result)
	})

	return
}

// gateway fetchs a gateway from the database without antennas and attributes and
// saves it into result.
func (s *GatewayStore) getByID(q db.QueryContext, gtwID uuid.UUID) (*ttnpb.Gateway, error) {
	result := new(ttnpb.Gateway)
	err := q.SelectOne(
		result,
		`SELECT
					gateway_id,
					description,
					frequency_plan_id,
					activated_at,
					privacy_settings,
					auto_update,
					platform,
					cluster_address
				FROM gateways
				WHERE id = $1`,
		gtwID)
	if db.IsNoRows(err) {
		return nil, ErrGatewayNotFound.New(nil)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update updates the gateway.
func (s *GatewayStore) Update(gateway store.Gateway) error {
	err := s.transact(func(tx *db.Tx) error {
		gtw := gateway.GetGateway()

		gtwID, err := s.getGatewayID(tx, gtw.GatewayIdentifiers)
		if err != nil {
			return err
		}

		err = s.update(tx, gtwID, gtw)
		if err != nil {
			return err
		}

		err = s.updateAttributes(tx, gtwID, gtw.Attributes)
		if err != nil {
			return err
		}

		err = s.updateAntennas(tx, gtwID, gtw.Antennas)
		if err != nil {
			return err
		}

		err = s.updateRadios(tx, gtwID, gtw.Radios)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, gtwID, gateway)
	})
	return err
}

func (s *GatewayStore) update(q db.QueryContext, gtwID uuid.UUID, data *ttnpb.Gateway) (err error) {
	var id string
	err = q.NamedSelectOne(
		&id,
		`UPDATE gateways
			SET
				description = :description,
				frequency_plan_id = :frequency_plan_id,
				activated_at = :activated_at,
				privacy_settings = :privacy_settings,
				auto_update = :auto_update,
				platform = :platform,
				cluster_address = :cluster_address,
				updated_at = current_timestamp()
			WHERE id = :id
			RETURNING gateway_id`,
		gateway{
			ID:      gtwID,
			Gateway: *data,
		})

	if db.IsNoRows(err) {
		err = ErrGatewayNotFound.New(nil)
	}

	return
}

func (s *GatewayStore) updateAntennas(q db.QueryContext, gtwID uuid.UUID, antennas []ttnpb.GatewayAntenna) error {
	_, err := q.Exec("DELETE FROM gateways_antennas WHERE gateway_id = $1", gtwID)
	if err != nil {
		return err
	}

	return s.addAntennas(q, gtwID, antennas)
}

func (s *GatewayStore) updateRadios(q db.QueryContext, gtwID uuid.UUID, radios []ttnpb.GatewayRadio) error {
	_, err := q.Exec("DELETE FROM gateways_radios WHERE gateway_id = $1", gtwID)
	if err != nil {
		return err
	}

	return s.addRadios(q, gtwID, radios)
}

// updateAttributes removes the attributes that no longer exists for the gateway
// given its ID and sets the rest of attributes.
func (s *GatewayStore) updateAttributes(q db.QueryContext, gtwID uuid.UUID, attributes map[string]string) error {
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
func (s *GatewayStore) removeAttributeDiffQuery(gtwID uuid.UUID, attributes map[string]string) (string, []interface{}) {
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
func (s *GatewayStore) setAttributes(q db.QueryContext, gtwID uuid.UUID, attributes map[string]string) error {
	if attributes == nil || len(attributes) == 0 {
		return nil
	}

	query, args := s.setAttributesQuery(gtwID, attributes)
	_, err := q.Exec(query, args...)
	return err
}

// setAttributesQuery is a helper that constructs the upsert query for the
// setAttributes method and returns it together with the list of arguments.
func (s *GatewayStore) setAttributesQuery(gtwID uuid.UUID, attributes map[string]string) (string, []interface{}) {
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
		j++
	}

	query := fmt.Sprintf(
		`UPSERT
			INTO gateways_attributes (gateway_id, attribute, value)
			VALUES %s`,
		strings.Join(boundValues, ", "))

	return query, args
}

func (s *GatewayStore) listAttributes(q db.QueryContext, gtwID uuid.UUID) (map[string]string, error) {
	var attrs []struct {
		Attribute string
		Value     string
	}
	err := q.Select(
		&attrs,
		`SELECT
				attribute,
				value
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

func (s *GatewayStore) listAntennas(q db.QueryContext, gtwID uuid.UUID) ([]ttnpb.GatewayAntenna, error) {
	var antnns []struct {
		Longitude float32
		Latitude  float32
		Altitude  int32
		ttnpb.GatewayAntenna
	}
	err := q.Select(
		&antnns,
		`SELECT
				longitude,
				latitude,
				altitude,
				gain,
				type,
				model,
				placement
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

func (s *GatewayStore) listRadios(q db.QueryContext, gtwID uuid.UUID) ([]ttnpb.GatewayRadio, error) {
	var radios []ttnpb.GatewayRadio
	err := q.Select(
		&radios,
		`SELECT
				frequency,
				tx_configuration
			FROM gateways_radios
			WHERE gateway_id = $1
			ORDER BY created_at ASC`,
		gtwID)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}
	if radios == nil {
		return make([]ttnpb.GatewayRadio, 0), nil
	}
	return radios, nil
}

// Delete deletes a gateway.
func (s *GatewayStore) Delete(ids ttnpb.GatewayIdentifiers) error {
	err := s.transact(func(tx *db.Tx) error {
		gtwID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		err = s.deleteCollaborators(tx, gtwID)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeys(tx, gtwID)
		if err != nil {
			return err
		}

		err = s.removeAntennas(tx, gtwID)
		if err != nil {
			return err
		}

		err = s.removeRadios(tx, gtwID)
		if err != nil {
			return err
		}

		err = s.removeAttributes(tx, gtwID)
		if err != nil {
			return err
		}

		return s.delete(tx, gtwID)
	})

	return err
}

// delete deletes the gateway itself. All rows in other tables that references
// this entity must be delete before this one gets deleted.
func (s *GatewayStore) delete(q db.QueryContext, gtwID uuid.UUID) error {
	id := new(string)
	err := q.SelectOne(
		id,
		`DELETE
			FROM gateways
			WHERE id = $1
			RETURNING gateway_id`,
		gtwID)
	if db.IsNoRows(err) {
		return ErrGatewayNotFound.New(nil)
	}
	return err
}

// removeAntennas removes all the antennas from a gateway.
func (s *GatewayStore) removeAntennas(q db.QueryContext, gtwID uuid.UUID) error {
	_, err := q.Exec("DELETE FROM gateways_antennas WHERE gateway_id = $1", gtwID)
	return err
}

// removeRadios removes all the radios from a gateway.
func (s *GatewayStore) removeRadios(q db.QueryContext, gtwID uuid.UUID) error {
	_, err := q.Exec("DELETE FROM gateways_radios WHERE gateway_id = $1", gtwID)
	return err
}

// removeAttributes removes all the attributes from a gateway.
func (s *GatewayStore) removeAttributes(q db.QueryContext, gtwID uuid.UUID) error {
	_, err := q.Exec("DELETE FROM gateways_attributes WHERE gateway_id = $1", gtwID)
	return err
}
