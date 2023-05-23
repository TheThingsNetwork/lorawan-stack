// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

// Package istelemetry contains telemetry functions regarding the collection of data in the IdentityServer.
package istelemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	store "go.thethings.network/lorawan-stack/v3/pkg/identityserver/bunstore"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	telemetry "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter/models"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
)

// EntityCountTaskName is the name of the task that collects entity counts.
// This is used as the task name in the TaskQueue's callback registry.
const EntityCountTaskName = "entity-count-task"

type isTelemetry struct {
	uid        string
	target     string
	httpClient *http.Client
	DB         *bun.DB
}

// Option to apply at istelemetry initialization.
type Option interface {
	apply(*isTelemetry)
}

type optionFunc func(*isTelemetry)

func (f optionFunc) apply(it *isTelemetry) { f(it) }

// WithUID sets the uid of the task.
func WithUID(uid string) Option {
	return optionFunc(func(it *isTelemetry) {
		it.uid = uid
	})
}

// WithTarget sets the URL to which the content will be sent to.
func WithTarget(target string) Option {
	return optionFunc(func(it *isTelemetry) {
		it.target = target
	})
}

// WithBunDB sets the DB to be used for the queries.
func WithBunDB(db *bun.DB) Option {
	return optionFunc(func(it *isTelemetry) {
		it.DB = db
	})
}

// WithHTTPClient sets the HTTP client to be used for the requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return optionFunc(func(it *isTelemetry) {
		it.httpClient = httpClient
	})
}

// Task is the interface that wraps the EntitiesCountTask method.
type Task interface {
	// Validate determines if the necessary requirement to run telemetry tasks are set.
	Validate(ctx context.Context) error
	// CountEntities fetches the number of entities in the database and sends it to the target.
	CountEntities(ctx context.Context) error
}

// New returns a instance of the identity server telemetry task.
func New(opts ...Option) Task {
	it := &isTelemetry{}
	for _, opt := range opts {
		opt.apply(it)
	}
	return it
}

func (it *isTelemetry) countApplications(ctx context.Context) (uint64, error) {
	n, err := it.DB.NewSelect().Model(&store.Application{}).Count(ctx)
	return uint64(n), storeutil.WrapDriverError(err)
}

func (it *isTelemetry) countEndDevices(ctx context.Context) (uint64, error) {
	n, err := it.DB.NewSelect().Model(&store.EndDevice{}).Count(ctx)
	return uint64(n), storeutil.WrapDriverError(err)
}

func (it *isTelemetry) countGateways(ctx context.Context) (uint64, error) {
	n, err := it.DB.NewSelect().Model(&store.Gateway{}).Count(ctx)
	return uint64(n), storeutil.WrapDriverError(err)
}

func (it *isTelemetry) countOrganizations(ctx context.Context) (resp models.OrganizationsCount, err error) {
	for _, q := range []struct {
		query *bun.SelectQuery
		num   *uint64
	}{
		{
			query: it.DB.NewSelect().Model(&store.Organization{}),
			num:   &resp.Total,
		},
	} {
		n, err := q.query.Count(ctx)
		if err != nil {
			return resp, storeutil.WrapDriverError(err)
		}
		*q.num = uint64(n)
	}
	return resp, nil
}

// countActiveDevices returns the ActivatedEndDevicesAmount, which counts active devices by day, week and month.
func (it *isTelemetry) countActiveDevices(ctx context.Context) (resp models.ActivateEndDevicesCount, err error) {
	for _, q := range []struct {
		query *bun.SelectQuery
		num   *uint64
	}{
		{
			query: it.DB.NewSelect().Model(&store.EndDevice{}).Where("activated_at IS NOT NULL"),
			num:   &resp.Total,
		},
		{
			query: it.DB.NewSelect().Model(&store.EndDevice{}).Where("activated_at > NOW() - INTERVAL '1 DAY'"),
			num:   &resp.LastDay,
		},
		{
			query: it.DB.NewSelect().Model(&store.EndDevice{}).Where("activated_at > NOW() - INTERVAL '1 WEEK'"),
			num:   &resp.LastWeek,
		},
		{
			query: it.DB.NewSelect().Model(&store.EndDevice{}).Where("activated_at > NOW() - INTERVAL '1 MONTH'"),
			num:   &resp.LastMonth,
		},
	} {
		n, err := q.query.Count(ctx)
		if err != nil {
			return resp, storeutil.WrapDriverError(err)
		}
		*q.num = uint64(n)
	}
	return resp, nil
}

func (it *isTelemetry) countGatewaysByFreqPlan(ctx context.Context) (map[string]uint64, error) {
	type result struct {
		FrequencyPlanID string `bun:"frequency_plan_id"`
		Count           uint64 `bun:"count"`
	}
	var results []result
	err := it.DB.NewSelect().
		Model(&store.Gateway{}).
		ColumnExpr("frequency_plan_id").
		ColumnExpr("COUNT(frequency_plan_id)").
		Order("frequency_plan_id").
		Group("frequency_plan_id").Scan(ctx, &results)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	m := make(map[string]uint64)
	for _, r := range results {
		for _, freqPlan := range strings.Split(r.FrequencyPlanID, " ") {
			m[freqPlan] += r.Count
		}
	}
	return m, nil
}

func (it *isTelemetry) countUserByTypes(ctx context.Context) (resp models.UsersCount, err error) {
	for _, q := range []struct {
		query *bun.SelectQuery
		num   *uint64
	}{
		{
			query: it.DB.NewSelect().Model(&store.User{}),
			num:   &resp.Total,
		},
		{
			query: it.DB.NewSelect().Model(&store.User{}).Where("admin IS TRUE"),
			num:   &resp.Admin,
		},
		{
			query: it.DB.NewSelect().Model(&store.User{}).Where("admin IS FALSE"),
			num:   &resp.Standard,
		},
	} {
		n, err := q.query.Count(ctx)
		if err != nil {
			return resp, storeutil.WrapDriverError(err)
		}
		*q.num = uint64(n)
	}
	return resp, nil
}

var errInsufficientConfiguration = errors.DefineFailedPrecondition(
	"insufficient_configuration", "insufficient configuration to start task, missing `{field}`",
)

// Validate if the necessary fields for the tasks are set.
func (it *isTelemetry) Validate(_ context.Context) error {
	if it.uid == "" {
		return errInsufficientConfiguration.WithAttributes("field", "uid")
	}
	if it.DB == nil {
		return errInsufficientConfiguration.WithAttributes("field", "db")
	}
	if it.httpClient == nil {
		return errInsufficientConfiguration.WithAttributes("field", "http_client")
	}
	if it.target == "" {
		return errInsufficientConfiguration.WithAttributes("field", "target")
	}
	return nil
}

// CountEntities is the task that collects data regarding the amount of each entity in the IS database.
func (it *isTelemetry) CountEntities(ctx context.Context) error {
	logger := log.FromContext(ctx)

	apps, err := it.countApplications(ctx)
	if err != nil {
		return err
	}
	devs, err := it.countEndDevices(ctx)
	if err != nil {
		return err
	}
	gtws, err := it.countGateways(ctx)
	if err != nil {
		return err
	}
	gtwsByFreqID, err := it.countGatewaysByFreqPlan(ctx)
	if err != nil {
		return err
	}
	orgs, err := it.countOrganizations(ctx)
	if err != nil {
		return err
	}
	activeDevs, err := it.countActiveDevices(ctx)
	if err != nil {
		return err
	}
	usrs, err := it.countUserByTypes(ctx)
	if err != nil {
		return err
	}

	data := &models.TelemetryMessage{
		UID: it.uid,
		OS:  telemetry.OSTelemetryData(),
		EntitiesCount: &models.EntitiesCount{
			Gateways: models.GatewaysCount{
				Total:                     gtws,
				GatewaysByFrequencyPlanID: gtwsByFreqID,
			},
			EndDevices: models.EndDevicesCount{
				Total:              devs,
				ActivateEndDevices: activeDevs,
			},
			Applications: models.ApplicationsCount{
				Total: apps,
			},
			Accounts: models.AccountsCount{
				Users:         usrs,
				Organizations: orgs,
			},
		},
	}
	logger.WithField("message", data).Debug("Collected entity count telemetry data")

	b, err := json.Marshal(data)
	if err != nil {
		logger.WithError(err).Debug("Failed to marshal telemetry information")
		return err
	}

	resp, err := it.httpClient.Post(it.target, "application/json", bytes.NewBuffer(b))
	if err != nil {
		logger.WithError(err).Debug("Failed to send information to telemetry server")
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}
