// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// OrganizationStore implements store.OrganizationStore.
type OrganizationStore struct {
	storer
	*extraAttributesStore
	*apiKeysStore
	*accountStore
}

func NewOrganizationStore(store storer) *OrganizationStore {
	return &OrganizationStore{
		storer:               store,
		extraAttributesStore: newExtraAttributesStore(store, "organization"),
		apiKeysStore:         newAPIKeysStore(store, "organization"),
		accountStore:         newAccountStore(store),
	}
}

// Create creates an organization.
func (s *OrganizationStore) Create(organization store.Organization) error {
	err := s.transact(func(tx *db.Tx) error {
		organizationID := organization.GetOrganization().OrganizationID

		err := s.accountStore.registerOrganizationID(tx, organizationID)
		if err != nil {
			return err
		}

		err = s.create(tx, organization)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, organizationID, organization, nil)
	})
	return err
}

func (s *OrganizationStore) create(q db.QueryContext, organization store.Organization) error {
	_, err := q.NamedExec(
		`INSERT
			INTO organizations (
				organization_id,
				name,
				description,
				url,
				location,
				email
			)
			VALUES (
				lower(:organization_id),
				:name,
				:description,
				:url,
				:location,
				lower(:email))`,
		organization.GetOrganization())
	return err
}

// GetByID finds the organization by ID and retrieves it.
func (s *OrganizationStore) GetByID(organizationID string, factory store.OrganizationSpecializer) (result store.Organization, err error) {
	err = s.transact(func(tx *db.Tx) error {
		organization, err := s.getByID(tx, organizationID)
		if err != nil {
			return err
		}

		result = factory(*organization)

		return s.loadAttributes(tx, organizationID, result)
	})

	return
}

func (s *OrganizationStore) getByID(q db.QueryContext, organizationID string) (*ttnpb.Organization, error) {
	result := new(ttnpb.Organization)
	err := q.SelectOne(
		result,
		`SELECT
				organization_id,
				name,
				description,
				url,
				location,
				email
			FROM organizations
			WHERE organization_id = $1`,
		organizationID)
	if db.IsNoRows(err) {
		return nil, ErrOrganizationNotFound.New(errors.Attributes{
			"organization_id": organizationID,
		})
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListByUser returns the organizations to which an user is a member of.
func (s *OrganizationStore) ListByUser(userID string, factory store.OrganizationSpecializer) (result []store.Organization, err error) {
	err = s.transact(func(tx *db.Tx) error {
		organizations, err := s.userOrganizations(tx, userID)
		if err != nil {
			return err
		}

		for _, organization := range organizations {
			org := factory(organization)

			err := s.loadAttributes(tx, organization.OrganizationID, org)
			if err != nil {
				return err
			}

			result = append(result, org)
		}

		return nil
	})

	return
}

func (s *OrganizationStore) userOrganizations(q db.QueryContext, userID string) ([]ttnpb.Organization, error) {
	var organizations []ttnpb.Organization
	err := q.Select(
		&organizations,
		`SELECT
				DISTINCT organization_id,
				name,
				description,
				url,
				location,
				email
			FROM organizations
			INNER JOIN organizations_members
			USING (organization_id)
			WHERE user_id = $1`,
		userID)
	if err != nil {
		return nil, err
	}
	return organizations, nil
}

func (s *OrganizationStore) Update(organization store.Organization) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.update(tx, organization)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, organization.GetOrganization().OrganizationID, organization, nil)
	})

	return err
}

func (s *OrganizationStore) update(q db.QueryContext, organization store.Organization) error {
	org := organization.GetOrganization()

	var id string
	err := q.NamedSelectOne(
		&id,
		`UPDATE organizations
			SET
				name = :name,
				description = :description,
				url = :url,
				location = :location,
				email = lower(:email),
				updated_at = current_timestamp()
			WHERE organization_id = :organization_id
			RETURNING organization_id`,
		org)

	if db.IsNoRows(err) {
		return ErrOrganizationNotFound.New(errors.Attributes{
			"organization_id": org.OrganizationID,
		})
	}

	return err
}

func (s *OrganizationStore) Delete(organizationID string) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.deleteCollaborations(tx, organizationID)
		if err != nil {
			return err
		}

		err = s.deleteMembers(tx, organizationID)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeys(tx, organizationID)
		if err != nil {
			return err
		}

		err = s.delete(tx, organizationID)
		if err != nil {
			return err
		}

		return s.accountStore.deleteID(tx, organizationID)
	})

	return err
}

func (s *OrganizationStore) deleteCollaborations(q db.QueryContext, organizationID string) error {
	_, err := q.Exec(
		`DELETE
				FROM applications_collaborators
				WHERE account_id = $1`,
		organizationID)
	if err != nil {
		return err
	}

	_, err = q.Exec(
		`DELETE
				FROM gateways_collaborators
				WHERE account_id = $1`,
		organizationID)
	return err
}

func (s *OrganizationStore) deleteMembers(q db.QueryContext, organizationID string) error {
	_, err := q.Exec(
		`DELETE
				FROM organizations_members
				WHERE organization_id = $1`,
		organizationID)
	return err
}

func (s *OrganizationStore) delete(q db.QueryContext, organizationID string) error {
	var id string
	err := q.SelectOne(
		&id,
		`DELETE
			FROM organizations
			WHERE organization_id = $1
			RETURNING organization_id`,
		organizationID)
	if db.IsNoRows(err) {
		return ErrOrganizationNotFound.New(errors.Attributes{
			"organization_id": id,
		})
	}
	return err
}

// HasMemberRights checks whether an user has or not a set of given rights to
// an organization. Returns false if the user is not part of the organization.
func (s *OrganizationStore) HasMemberRights(organizationID, userID string, rights ...ttnpb.Right) (bool, error) {
	return s.hasMemberRights(s.queryer(), organizationID, userID, rights...)
}

func (s *OrganizationStore) hasMemberRights(q db.QueryContext, organizationID, userID string, rights ...ttnpb.Right) (bool, error) {
	clauses := make([]string, 0, len(rights))
	args := make([]interface{}, 0, len(rights)+2)
	args = append(args, organizationID, userID)

	for i, right := range rights {
		args = append(args, right)
		clauses = append(clauses, fmt.Sprintf(`"right" = $%d`, i+3))
	}

	found := 0
	err := q.SelectOne(
		&found,
		fmt.Sprintf(`
			SELECT
					COUNT(user_id)
				FROM organizations_members
				WHERE organization_id = $1 AND user_id = $2 AND (%s)`, strings.Join(clauses, " OR ")),
		args...)
	if db.IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return (found == len(rights)), nil
}

// ListMembers retrieves all the members from an organization. Optionally a list
// of rights can be passed to filter them.
func (s *OrganizationStore) ListMembers(organizationID string, rights ...ttnpb.Right) ([]*ttnpb.OrganizationMember, error) {
	return s.listMembers(s.queryer(), organizationID, rights...)
}

func (s *OrganizationStore) listMembers(q db.QueryContext, organizationID string, rights ...ttnpb.Right) ([]*ttnpb.OrganizationMember, error) {
	query := ""
	args := make([]interface{}, 0, 2+len(rights))
	args = append(args, organizationID)

	if len(rights) == 0 {
		query = `
			SELECT
				organization_id,
				user_id,
				"right"
			FROM organizations_members
			WHERE organization_id = $1`
	} else {
		args = append(args, len(rights))
		clauses := make([]string, 0, len(rights))

		for i, right := range rights {
			clauses = append(clauses, fmt.Sprintf(`"right" = $%d`, i+3))
			args = append(args, right)
		}

		query = `
			SELECT
				organization_id,
				user_id,
				"right"
			FROM organizations_members
			WHERE organization_id = $1 AND user_id IN (
				SELECT
					user_id
				FROM (
					SELECT
						user_id,
						COUNT(user_id) AS count
					FROM organizations_members
					WHERE organization_id = $1 AND (%s)
					GROUP BY user_id
				)
				WHERE count = $2
			)`

		query = fmt.Sprintf(query, strings.Join(clauses, " OR "))
	}

	var rows []*struct {
		*ttnpb.OrganizationMember
		Right ttnpb.Right
	}
	err := q.Select(&rows, query, args...)
	if err != nil {
		return nil, err
	}
	if rows == nil || len(rows) == 0 {
		return make([]*ttnpb.OrganizationMember, 0), nil
	}

	// map the rows by User ID
	byUser := make(map[string]*ttnpb.OrganizationMember)
	for _, row := range rows {
		_, ok := byUser[row.OrganizationMember.UserID]
		if !ok {
			byUser[row.UserID] = new(ttnpb.OrganizationMember)
			byUser[row.UserID].UserIdentifier = ttnpb.UserIdentifier{row.UserID}
			byUser[row.UserID].OrganizationIdentifier = ttnpb.OrganizationIdentifier{row.OrganizationID}
			byUser[row.UserID].Rights = make([]ttnpb.Right, 0, 1)
		}

		byUser[row.UserID].Rights = append(byUser[row.UserID].Rights, row.Right)
	}

	members := make([]*ttnpb.OrganizationMember, 0, len(byUser))
	for _, member := range byUser {
		members = append(members, member)
	}

	return members, nil
}

func (s *OrganizationStore) SetMember(member *ttnpb.OrganizationMember) error {
	if len(member.Rights) == 0 {
		return s.removeMember(s.queryer(), member)
	}
	err := s.transact(func(tx *db.Tx) error {
		return s.setMember(tx, member)
	})
	return err
}

func (s *OrganizationStore) removeMember(q db.QueryContext, member *ttnpb.OrganizationMember) error {
	_, err := q.Exec(
		`DELETE FROM
			organizations_members
			WHERE organization_id = $1 AND user_id = $2`,
		member.OrganizationID,
		member.UserID)
	return err
}

func (s *OrganizationStore) setMember(q db.QueryContext, member *ttnpb.OrganizationMember) error {
	// remove rights
	err := s.removeMember(q, member)
	if err != nil {
		return err
	}

	// add rights
	values := make([]string, len(member.Rights))
	args := make([]interface{}, len(member.Rights)+2)
	args[0] = member.OrganizationID
	args[1] = member.UserID
	for i, right := range member.Rights {
		values[i] = fmt.Sprintf("($1, $2, $%d)", i+3)
		args[i+2] = right
	}

	query := fmt.Sprintf(
		`INSERT
			INTO organizations_members (organization_id, user_id, "right")
			VALUES %s
			ON CONFLICT (organization_id, user_id, "right")
			DO NOTHING`,
		strings.Join(values, ", "))

	_, err = q.Exec(query, args...)
	return err
}

// ListUserRights returns the rights a given user has for an entity.
func (s *OrganizationStore) ListUserRights(organizationID string, userID string) ([]ttnpb.Right, error) {
	return s.listUserRights(s.queryer(), organizationID, userID)
}

func (s *OrganizationStore) listUserRights(q db.QueryContext, organizationID string, userID string) ([]ttnpb.Right, error) {
	var rights []ttnpb.Right
	err := q.Select(
		&rights,
		`SELECT
				"right"
			FROM organizations_members
			WHERE organization_id = $1 AND user_id = $2`,
		organizationID,
		userID)
	if err != nil {
		return nil, err
	}
	return rights, nil
}

// LoadAttributes loads the extra attributes in organization if it is a store.Attributer.
func (s *OrganizationStore) LoadAttributes(organizationID string, organization store.Organization) error {
	return s.loadAttributes(s.queryer(), organizationID, organization)
}

func (s *OrganizationStore) loadAttributes(q db.QueryContext, organizationID string, organization store.Organization) error {
	attr, ok := organization.(store.Attributer)
	if ok {
		return s.extraAttributesStore.loadAttributes(q, organizationID, attr)
	}

	return nil
}

// StoreAttributes store the extra attributes of organization if it is a
// store.Attributer and writes the resulting organization in result.
func (s *OrganizationStore) StoreAttributes(organizationID string, organization, result store.Organization) error {
	return s.storeAttributes(s.queryer(), organizationID, organization, result)
}

func (s *OrganizationStore) storeAttributes(q db.QueryContext, organizationID string, organization, result store.Organization) error {
	attr, ok := organization.(store.Attributer)
	if ok {
		res, ok := result.(store.Attributer)
		if result == nil || !ok {
			return s.extraAttributesStore.storeAttributes(q, organizationID, attr, nil)
		}

		return s.extraAttributesStore.storeAttributes(q, organizationID, attr, res)
	}

	return nil
}
