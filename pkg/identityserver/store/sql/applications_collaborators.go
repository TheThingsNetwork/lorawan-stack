// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sql

import (
	"fmt"
	"strings"

	"github.com/satori/go.uuid"
	"go.thethings.network/lorawan-stack/pkg/identityserver/db"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// ListByOrganizationOrUser returns the applications to which an organization or
// user if collaborator of.
func (s *applicationStore) ListByOrganizationOrUser(ids ttnpb.OrganizationOrUserIdentifiers, specializer store.ApplicationSpecializer) (result []store.Application, err error) {
	err = s.transact(func(tx *db.Tx) error {
		accountID, err := s.getAccountID(tx, ids)
		if err != nil {
			return err
		}

		applications, err := s.listOrganizationOrUserApplications(tx, accountID)
		if err != nil {
			return err
		}

		for _, application := range applications {
			specialized := specializer(application.Application)

			err := s.loadAttributes(tx, application.ID, specialized)
			if err != nil {
				return err
			}

			result = append(result, specialized)
		}

		return nil
	})
	return
}

func (s *applicationStore) listOrganizationOrUserApplications(q db.QueryContext, accountID uuid.UUID) (applications []application, err error) {
	err = q.Select(
		&applications,
		`SELECT DISTINCT applications.*
			FROM applications
			JOIN applications_collaborators
			ON (
				applications.id = applications_collaborators.application_id
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

// SetCollaborator sets a collaborator into an application.
// If the provided list of rights is empty the collaborator will be unset.
func (s *applicationStore) SetCollaborator(collaborator ttnpb.ApplicationCollaborator) error {
	err := s.transact(func(tx *db.Tx) error {
		appID, err := s.getApplicationID(tx, collaborator.ApplicationIdentifiers)
		if err != nil {
			return err
		}

		accountID, err := s.getAccountID(tx, collaborator.OrganizationOrUserIdentifiers)
		if err != nil {
			return err
		}

		err = s.unsetCollaborator(tx, appID, accountID)
		if err != nil {
			return err
		}

		if len(collaborator.Rights) == 0 {
			return nil
		}

		return s.setCollaborator(tx, appID, accountID, collaborator.Rights)
	})
	return err
}

func (s *applicationStore) unsetCollaborator(q db.QueryContext, appID, accountID uuid.UUID) error {
	_, err := q.Exec(
		`DELETE
			FROM applications_collaborators
			WHERE application_id = $1 AND account_id = $2`,
		appID,
		accountID)
	return err
}

func (s *applicationStore) setCollaborator(q db.QueryContext, appID, accountID uuid.UUID, rights []ttnpb.Right) (err error) {
	args := make([]interface{}, 0, 2+len(rights))
	args = append(args, appID, accountID)

	boundValues := make([]string, 0, len(rights))
	for i, right := range rights {
		args = append(args, right)
		boundValues = append(boundValues, fmt.Sprintf("($1, $2, $%d)", i+3))
	}

	query := fmt.Sprintf(
		`INSERT
			INTO applications_collaborators (application_id, account_id, "right")
			VALUES %s
			ON CONFLICT (application_id, account_id, "right")
			DO NOTHING`,
		strings.Join(boundValues, ", "))

	_, err = q.Exec(query, args...)

	return err
}

// HasCollaboratorRights checks whether a collaborator has a given set of rights
// to an application. It returns false if the collaborationship does not exist.
func (s *applicationStore) HasCollaboratorRights(id ttnpb.ApplicationIdentifiers, target ttnpb.OrganizationOrUserIdentifiers, rights ...ttnpb.Right) (result bool, err error) {
	err = s.transact(func(tx *db.Tx) error {
		appID, err := s.getApplicationID(tx, id)
		if err != nil {
			return err
		}

		accountID, err := s.getAccountID(tx, target)
		if err != nil {
			return err
		}

		result, err = s.hasCollaboratorRights(tx, appID, accountID, rights...)
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *applicationStore) hasCollaboratorRights(q db.QueryContext, appID, accountID uuid.UUID, rights ...ttnpb.Right) (bool, error) {
	clauses := make([]string, 0, len(rights))
	args := make([]interface{}, 0, len(rights)+1)
	args = append(args, appID, accountID)

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
				FROM applications_collaborators
				WHERE (%s) AND application_id = $1 AND (account_id = $2 OR account_id IN (
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

// ListCollaborators retrieves all the collaborators from an application.
// Optionally a list of rights can be passed as argument to filter them.
func (s *applicationStore) ListCollaborators(ids ttnpb.ApplicationIdentifiers, rights ...ttnpb.Right) (collaborators []ttnpb.ApplicationCollaborator, err error) {
	err = s.transact(func(tx *db.Tx) error {
		appID, err := s.getApplicationID(tx, ids)
		if err != nil {
			return err
		}

		collaborators, err = s.listCollaborators(tx, appID, rights...)
		if err != nil {
			return err
		}

		for i := range collaborators {
			collaborators[i].ApplicationIdentifiers = ids
		}

		return nil
	})
	return
}

// nolint: dupl
func (s *applicationStore) listCollaborators(q db.QueryContext, appID uuid.UUID, rights ...ttnpb.Right) ([]ttnpb.ApplicationCollaborator, error) {
	args := make([]interface{}, 1)
	args[0] = appID

	var query string
	if len(rights) == 0 {
		query = `
		SELECT
			applications_collaborators.account_id,
			"right",
			type
		FROM applications_collaborators
		JOIN accounts ON (accounts.id = applications_collaborators.account_id)
		WHERE application_id = $1`
	} else {
		rightsClause := make([]string, 0, len(rights))
		for _, right := range rights {
			rightsClause = append(rightsClause, fmt.Sprintf(`"right" = '%d'`, right))
		}

		query = fmt.Sprintf(`
			SELECT
					applications_collaborators.account_id,
					"right",
					type
	    	FROM applications_collaborators
	    	JOIN accounts ON (accounts.id = applications_collaborators.account_id)
	    	WHERE application_id = $1 AND applications_collaborators.account_id IN
	    	(
	      	SELECT
	      			account_id
	      		FROM
	      			(
	          		SELECT
	          				account_id,
	          				count(account_id) as count
	          	  	FROM applications_collaborators
	          			WHERE application_id = $1 AND (%s)
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

	byUser := make(map[string]*ttnpb.ApplicationCollaborator)
	for _, collaborator := range collaborators {
		if _, exists := byUser[collaborator.AccountID.String()]; !exists {
			var identifier ttnpb.OrganizationOrUserIdentifiers
			if collaborator.Type == organizationIDType {
				id, err := s.store().Organizations.getOrganizationIdentifiersFromID(q, collaborator.AccountID)
				if err != nil {
					return nil, err
				}

				identifier = ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_OrganizationID{OrganizationID: &id}}
			} else {
				id, err := s.store().Users.getUserIdentifiersFromID(q, collaborator.AccountID)
				if err != nil {
					return nil, err
				}

				identifier = ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &id}}
			}

			byUser[collaborator.AccountID.String()] = &ttnpb.ApplicationCollaborator{
				OrganizationOrUserIdentifiers: identifier,
				Rights: []ttnpb.Right{collaborator.Right},
			}
			continue
		}

		byUser[collaborator.AccountID.String()].Rights = append(byUser[collaborator.AccountID.String()].Rights, collaborator.Right)
	}

	result := make([]ttnpb.ApplicationCollaborator, 0, len(byUser))
	for _, collaborator := range byUser {
		result = append(result, *collaborator)
	}

	return result, nil
}

// ListCollaboratorRights returns the rights a given collaborator has for an
// Application. Returns empty list if the collaborationship does not exist.
func (s *applicationStore) ListCollaboratorRights(ids ttnpb.ApplicationIdentifiers, target ttnpb.OrganizationOrUserIdentifiers) (rights []ttnpb.Right, err error) {
	err = s.transact(func(tx *db.Tx) error {
		appID, err := s.getApplicationID(tx, ids)
		if err != nil {
			return err
		}

		accountID, err := s.getAccountID(tx, target)
		if err != nil {
			return err
		}

		return s.listCollaboratorRights(tx, appID, accountID, &rights)
	})
	return
}

func (s *applicationStore) listCollaboratorRights(q db.QueryContext, appID, accountID uuid.UUID, result *[]ttnpb.Right) error {
	err := q.Select(
		result, `
		SELECT
			"right"
		FROM applications_collaborators
		WHERE application_id = $1
		AND ( account_id = $2
			OR account_id IN
				( SELECT organization_id
				FROM organizations_members
				WHERE user_id = $2 ) )`,
		appID,
		accountID)
	return err
}
