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

	"github.com/cloudbase/garm-provider-common/params"
	"github.com/stretchr/testify/require"
)

type testPriceEstimator struct {
	price params.InstancePrice
	err   error
	got   params.BootstrapInstance
}

func (t *testPriceEstimator) GetInstancePrice(_ context.Context, bootstrapParams params.BootstrapInstance) (params.InstancePrice, error) {
	t.got = bootstrapParams
	return t.price, t.err
}

func TestRunGetInstancePrice(t *testing.T) {
	estimator := &testPriceEstimator{
		price: params.InstancePrice{
			Currency:          "USD",
			HourlyPrice:       0.5,
			ProvisioningModel: params.ProvisioningModelSpot,
			Components:        []params.InstancePriceComponent{{Name: "vcpu", Quantity: 8, HourlyPrice: 0.5}},
		},
	}
	bootstrap := params.BootstrapInstance{Name: "runner", Flavor: "c4d-standard-8"}

	out, err := RunGetInstancePrice(context.Background(), estimator, bootstrap)
	require.NoError(t, err)
	require.Equal(t, "c4d-standard-8", estimator.got.Flavor)

	var price params.InstancePrice
	require.NoError(t, json.Unmarshal([]byte(out), &price))
	require.Equal(t, estimator.price.HourlyPrice, price.HourlyPrice)
	require.Equal(t, params.ProvisioningModelSpot, price.ProvisioningModel)
	require.Len(t, price.Components, 1)
}

func TestRunGetInstancePriceError(t *testing.T) {
	estimator := &testPriceEstimator{err: errors.New("boom")}
	_, err := RunGetInstancePrice(context.Background(), estimator, params.BootstrapInstance{Flavor: "x"})
	require.ErrorContains(t, err, "boom")
}

func TestRunGetInstancePriceNotSupported(t *testing.T) {
	_, err := RunGetInstancePrice(context.Background(), struct{}{}, params.BootstrapInstance{Flavor: "x"})
	require.ErrorContains(t, err, "does not support GetInstancePrice")
}

func TestValidatePriceParams(t *testing.T) {
	require.Error(t, ValidatePriceParams(params.BootstrapInstance{}))
	require.NoError(t, ValidatePriceParams(params.BootstrapInstance{Flavor: "x"}))
}
