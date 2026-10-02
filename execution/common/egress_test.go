// Copyright 2026 Cloudbase Solutions SRL
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package common

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/cloudbase/garm-provider-common/params"
	"github.com/stretchr/testify/require"
)

type testEgressEstimator struct {
	egress     params.InstanceEgress
	err        error
	instance   string
	start, end time.Time
}

func (t *testEgressEstimator) GetInstanceEgress(_ context.Context, instance string, start, end time.Time) (params.InstanceEgress, error) {
	t.instance, t.start, t.end = instance, start, end
	return t.egress, t.err
}

func TestRunGetInstanceEgress(t *testing.T) {
	t.Setenv(EgressStartEnv, "2026-10-02T01:00:00Z")
	t.Setenv(EgressEndEnv, "2026-10-02T11:00:00+09:00")
	estimator := &testEgressEstimator{egress: params.InstanceEgress{
		Currency: "USD",
		Flows:    []params.EgressFlow{{LocationType: "EXTERNAL", Continent: "America", Bytes: 1 << 30}},
		Charges:  []params.EgressCharge{{Destination: "Americas", Bytes: 1 << 30, UnitPrice: 0.12, Cost: 0.12}},
		Cost:     0.12,
	}}

	out, err := RunGetInstanceEgress(context.Background(), estimator, "us-east4-a/garm-runner")
	require.NoError(t, err)
	require.Equal(t, "us-east4-a/garm-runner", estimator.instance)
	require.True(t, estimator.start.Equal(time.Date(2026, 10, 2, 1, 0, 0, 0, time.UTC)))
	require.True(t, estimator.end.Equal(time.Date(2026, 10, 2, 2, 0, 0, 0, time.UTC)))

	var egress params.InstanceEgress
	require.NoError(t, json.Unmarshal([]byte(out), &egress))
	require.Equal(t, 0.12, egress.Cost)
	require.Len(t, egress.Flows, 1)
}

func TestRunGetInstanceEgressErrors(t *testing.T) {
	t.Setenv(EgressStartEnv, "2026-10-02T01:00:00Z")
	t.Setenv(EgressEndEnv, "2026-10-02T02:00:00Z")

	_, err := RunGetInstanceEgress(context.Background(), struct{}{}, "x")
	require.ErrorContains(t, err, "does not support GetInstanceEgress")

	_, err = RunGetInstanceEgress(context.Background(), &testEgressEstimator{err: errors.New("boom")}, "x")
	require.ErrorContains(t, err, "boom")

	t.Setenv(EgressEndEnv, "2026-10-02T00:00:00Z")
	_, err = RunGetInstanceEgress(context.Background(), &testEgressEstimator{}, "x")
	require.ErrorContains(t, err, "must be before")

	t.Setenv(EgressEndEnv, "tomorrow")
	_, err = RunGetInstanceEgress(context.Background(), &testEgressEstimator{}, "x")
	require.ErrorContains(t, err, EgressEndEnv)
}
