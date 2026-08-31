// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package shoot

import (
	"context"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
)

// SeedCostProvider resolves the relative cost of running a shoot control plane
// on a given seed.
type SeedCostProvider interface {
	// GetSeedCost returns the cost for the given seed. known is false when the
	// provider has no cost data for the seed; err is reserved for lookup
	// failures (network, etc.).
	GetSeedCost(ctx context.Context, seed *gardencorev1beta1.Seed) (cost int, known bool, err error)
}
