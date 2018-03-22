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

// deleteCollaborators deletes all the collaborators from one gateway.
func (s *GatewayStore) deleteCollaborators(q db.QueryContext, gtwID uuid.UUID) error {
	_, err := q.Exec(
		`DELETE
			FROM gateways_collaborators
			WHERE gateway_id = $1`,
		gtwID)
	return err
}

// ListByOrganizationOrUser returns all the gateways to which an organization
// or user is collaborator of.
func (s *GatewayStore) ListByOrganizationOrUser(ids ttnpb.OrganizationOrUserIdentifiers, specializer store.GatewaySpecializer) (result []store.Gateway, err error) {
	err = s.transact(func(tx *db.Tx) error {
		accountID, err := s.getAccountID(tx, ids)
		if err != nil {
			return err
		}

		gateways, err := s.listOrganizationOrUserGateways(tx, accountID)
		if err != nil {
			return err
		}

		for _, gateway := range gateways {
			specialized := specializer(gateway.Gateway)

			attributes, err := s.listAttributes(tx, gateway.ID)
			if err != nil {
				return err
			}
			specialized.SetAttributes(attributes)

			antennas, err := s.listAntennas(tx, gateway.ID)
			if err != nil {
				return err
			}
			specialized.SetAntennas(antennas)

			radios, err := s.listRadios(tx, gateway.ID)
			if err != nil {
				return err
			}
			specialized.SetRadios(radios)

			err = s.loadAttributes(tx, gateway.ID, specialized)
			if err != nil {
				return err
			}

			result = append(result, specialized)
		}

		return nil
	})
	return
}

func (s *GatewayStore) listOrganizationOrUserGateways(q db.QueryContext, accountID uuid.UUID) (gateways []gateway, err error) {
	err = q.Select(
		&gateways,
		`SELECT DISTINCT gateways.*
			FROM gateways
			JOIN gateways_collaborators
			ON (
				gateways.id = gateways_collaborators.gateway_id
				AND
				(
					account_id = $1
					OR
					account_id IN (
						SELECT
							organization_id
						FROM organizations_members
						WHERE user_id = $1
					)
				)
			)`,
		accountID)
	return
}

// SetCollaborator inserts or modifies a collaborator within an entity.
// If the provided list of rights is empty the collaborator will be unset.
func (s *GatewayStore) SetCollaborator(collaborator ttnpb.GatewayCollaborator) error {
	err := s.transact(func(tx *db.Tx) error {
		gtwID, err := s.getGatewayID(tx, collaborator.GatewayIdentifiers)
		if err != nil {
			return err
		}

		accountID, err := s.getAccountID(tx, collaborator.OrganizationOrUserIdentifiers)
		if err != nil {
			return err
		}

		err = s.unsetCollaborator(tx, gtwID, accountID)
		if err != nil {
			return err
		}

		if len(collaborator.Rights) == 0 {
			return nil
		}

		return s.setCollaborator(tx, gtwID, accountID, collaborator.Rights)
	})
	return err
}

func (s *GatewayStore) unsetCollaborator(q db.QueryContext, gtwID, accountID uuid.UUID) error {
	_, err := q.Exec(
		`DELETE
			FROM gateways_collaborators
			WHERE gateway_id = $1 AND account_id = $2`, gtwID, accountID)
	return err
}

func (s *GatewayStore) setCollaborator(q db.QueryContext, gtwID, accountID uuid.UUID, rights []ttnpb.Right) (err error) {
	args := make([]interface{}, 2+len(rights))
	args[0] = gtwID
	args[1] = accountID

	boundValues := make([]string, len(rights))

	for i, right := range rights {
		args[i+2] = right
		boundValues[i] = fmt.Sprintf("($1, $2, $%d)", i+3)
	}

	query := fmt.Sprintf(
		`INSERT
			INTO gateways_collaborators (gateway_id, account_id, "right")
			VALUES %s
			ON CONFLICT (gateway_id, account_id, "right")
			DO NOTHING`,
		strings.Join(boundValues, " ,"))

	_, err = q.Exec(query, args...)

	return err
}

// HasCollaboratorRights checks whether a collaborator has a given set of rights
// to a gateway. It returns false if the collaborationship does not exist.
func (s *GatewayStore) HasCollaboratorRights(ids ttnpb.GatewayIdentifiers, target ttnpb.OrganizationOrUserIdentifiers, rights ...ttnpb.Right) (result bool, err error) {
	err = s.transact(func(tx *db.Tx) error {
		gtwID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		accountID, err := s.getAccountID(tx, target)
		if err != nil {
			return err
		}

		result, err = s.hasCollaboratorRights(tx, gtwID, accountID, rights...)
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *GatewayStore) hasCollaboratorRights(q db.QueryContext, gtwID, accountID uuid.UUID, rights ...ttnpb.Right) (bool, error) {
	clauses := make([]string, 0, len(rights))
	args := make([]interface{}, 0, len(rights)+1)
	args = append(args, gtwID, accountID)

	for i, right := range rights {
		args = append(args, right)
		clauses = append(clauses, fmt.Sprintf(`"right" = $%d`, i+3))
	}

	count := 0
	err := q.SelectOne(
		&count,
		fmt.Sprintf(
			`SELECT
				COUNT(DISTINCT "right")
				FROM gateways_collaborators
				WHERE (%s) AND gateway_id = $1 AND (account_id = $2 OR account_id IN (
					SELECT
						organization_id
					FROM organizations_members
					WHERE user_id = $2
				))`, strings.Join(clauses, " OR ")),
		args...)
	if db.IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return len(rights) == count, nil
}

// ListCollaborators retrieves all the collaborators from an entity.
func (s *GatewayStore) ListCollaborators(ids ttnpb.GatewayIdentifiers, rights ...ttnpb.Right) (collaborators []ttnpb.GatewayCollaborator, err error) {
	err = s.transact(func(tx *db.Tx) error {
		gtwID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		collaborators, err = s.listCollaborators(tx, gtwID, rights...)
		if err != nil {
			return err
		}

		for i := range collaborators {
			collaborators[i].GatewayIdentifiers = ids
		}

		return nil
	})
	return
}

// nolint: dupl
func (s *GatewayStore) listCollaborators(q db.QueryContext, gtwID uuid.UUID, rights ...ttnpb.Right) ([]ttnpb.GatewayCollaborator, error) {
	args := make([]interface{}, 1)
	args[0] = gtwID

	var query string
	if len(rights) == 0 {
		query = `
		SELECT
			gateways_collaborators.account_id,
			"right",
			type
		FROM gateways_collaborators
		JOIN accounts ON (accounts.id = gateways_collaborators.account_id)
		WHERE gateway_id = $1`
	} else {
		rightsClause := make([]string, 0, len(rights))
		for _, right := range rights {
			rightsClause = append(rightsClause, fmt.Sprintf(`"right" = '%d'`, right))
		}

		query = fmt.Sprintf(`
			SELECT
					gateways_collaborators.account_id,
					"right",
					type
	    	FROM gateways_collaborators
	    	JOIN accounts ON (accounts.id = gateways_collaborators.account_id)
	    	WHERE gateway_id = $1 AND gateways_collaborators.account_id IN
	    	(
	      	SELECT account_id
	      		FROM
	      			(
	          		SELECT
	          				account_id,
	          				count(account_id) as count
	          	  	FROM gateways_collaborators
	          			WHERE gateway_id = $1 AND (%s)
	          			GROUP BY account_id
	      			)
	      		WHERE count = $2
	  		)`,
			strings.Join(rightsClause, " OR "))

		args = append(args, len(rights))
	}

	var collaborators []struct {
		Right     ttnpb.Right
		AccountID uuid.UUID
		Type      int
	}
	err := q.Select(&collaborators, query, args...)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	byUser := make(map[string]*ttnpb.GatewayCollaborator)
	for _, collaborator := range collaborators {
		if _, exists := byUser[collaborator.AccountID.String()]; !exists {
			var identifier ttnpb.OrganizationOrUserIdentifiers
			if collaborator.Type == organizationIDType {
				id, err := s.store().Organizations.(*OrganizationStore).getOrganizationIdentifiersFromID(q, collaborator.AccountID)
				if err != nil {
					return nil, err
				}

				identifier = ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_OrganizationID{OrganizationID: &id}}
			} else {
				id, err := s.store().Users.(*UserStore).getUserIdentifiersFromID(q, collaborator.AccountID)
				if err != nil {
					return nil, err
				}

				identifier = ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &id}}
			}

			byUser[collaborator.AccountID.String()] = &ttnpb.GatewayCollaborator{
				OrganizationOrUserIdentifiers: identifier,
				Rights: []ttnpb.Right{collaborator.Right},
			}
			continue
		}

		byUser[collaborator.AccountID.String()].Rights = append(byUser[collaborator.AccountID.String()].Rights, collaborator.Right)
	}

	result := make([]ttnpb.GatewayCollaborator, 0, len(byUser))
	for _, collaborator := range byUser {
		result = append(result, *collaborator)
	}

	return result, nil
}

// ListCollaboratorRights returns the rights a given collaborator has for an
// Gateway. Returns empty list if the collaborationship does not exist.
func (s *GatewayStore) ListCollaboratorRights(ids ttnpb.GatewayIdentifiers, target ttnpb.OrganizationOrUserIdentifiers) (rights []ttnpb.Right, err error) {
	err = s.transact(func(tx *db.Tx) error {
		gtwID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		accountID, err := s.getAccountID(tx, target)
		if err != nil {
			return err
		}

		return s.listCollaboratorRights(tx, gtwID, accountID, &rights)
	})
	return
}

func (s *GatewayStore) listCollaboratorRights(q db.QueryContext, gtwID, accountID uuid.UUID, result *[]ttnpb.Right) error {
	err := q.Select(
		result, `
		SELECT
			"right"
		FROM gateways_collaborators
		WHERE gateway_id = $1
		AND ( account_id = $2
			OR account_id IN
				( SELECT organization_id
				FROM organizations_members
				WHERE user_id = $2 ) )`,
		gtwID,
		accountID)
	return err
}
