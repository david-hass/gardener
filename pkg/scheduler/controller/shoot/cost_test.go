// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package shoot

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
)

// fakeCostProvider resolves seed costs from a static per-seed-name table. A missing entry reports the cost as unknown;
// an entry present in errs makes GetSeedCost return an error for that seed.
type fakeCostProvider struct {
	costs map[string]int
	errs  map[string]error
}

func (f fakeCostProvider) GetSeedCost(_ context.Context, seed *gardencorev1beta1.Seed) (int, bool, error) {
	if err, ok := f.errs[seed.Name]; ok {
		return 0, false, err
	}
	cost, known := f.costs[seed.Name]
	return cost, known, nil
}

var _ = Describe("determineCandidatesWithMinimalCostStrategy", func() {
	var (
		ctx   = context.Background()
		log   = logr.Discard()
		shoot = &gardencorev1beta1.Shoot{}
	)

	makeSeed := func(name string) gardencorev1beta1.Seed {
		return gardencorev1beta1.Seed{ObjectMeta: metav1.ObjectMeta{Name: name}}
	}

	seedNames := func(seeds []gardencorev1beta1.Seed) []string {
		names := make([]string, len(seeds))
		for i, s := range seeds {
			names[i] = s.Name
		}
		return names
	}

	It("returns the single cheapest seed", func() {
		seeds := []gardencorev1beta1.Seed{makeSeed("a"), makeSeed("b"), makeSeed("c")}
		provider := fakeCostProvider{costs: map[string]int{"a": 30, "b": 10, "c": 20}}

		result, err := determineCandidatesWithMinimalCostStrategy(ctx, log, shoot, seeds, provider)
		Expect(err).NotTo(HaveOccurred())
		Expect(seedNames(result)).To(ConsistOf("b"))
	})

	It("returns all seeds tying at the minimal cost", func() {
		seeds := []gardencorev1beta1.Seed{makeSeed("a"), makeSeed("b"), makeSeed("c")}
		provider := fakeCostProvider{costs: map[string]int{"a": 10, "b": 10, "c": 20}}

		result, err := determineCandidatesWithMinimalCostStrategy(ctx, log, shoot, seeds, provider)
		Expect(err).NotTo(HaveOccurred())
		Expect(seedNames(result)).To(ConsistOf("a", "b"))
	})

	It("keeps seeds with unknown cost only as a last resort", func() {
		seeds := []gardencorev1beta1.Seed{makeSeed("priced"), makeSeed("unpriced")}
		provider := fakeCostProvider{costs: map[string]int{"priced": 50}}

		result, err := determineCandidatesWithMinimalCostStrategy(ctx, log, shoot, seeds, provider)
		Expect(err).NotTo(HaveOccurred())
		Expect(seedNames(result)).To(ConsistOf("priced"))
	})

	It("keeps all seeds when none have known cost", func() {
		seeds := []gardencorev1beta1.Seed{makeSeed("a"), makeSeed("b")}
		provider := fakeCostProvider{}

		result, err := determineCandidatesWithMinimalCostStrategy(ctx, log, shoot, seeds, provider)
		Expect(err).NotTo(HaveOccurred())
		Expect(seedNames(result)).To(ConsistOf("a", "b"))
	})

	It("skips seeds whose cost cannot be resolved", func() {
		seeds := []gardencorev1beta1.Seed{makeSeed("a"), makeSeed("boom"), makeSeed("c")}
		provider := fakeCostProvider{
			costs: map[string]int{"a": 40, "c": 10},
			errs:  map[string]error{"boom": errors.New("lookup failed")},
		}

		result, err := determineCandidatesWithMinimalCostStrategy(ctx, log, shoot, seeds, provider)
		Expect(err).NotTo(HaveOccurred())
		Expect(seedNames(result)).To(ConsistOf("c"))
	})

	It("returns the seed list unchanged when the provider is nil", func() {
		seeds := []gardencorev1beta1.Seed{makeSeed("a"), makeSeed("b")}

		result, err := determineCandidatesWithMinimalCostStrategy(ctx, log, shoot, seeds, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(seedNames(result)).To(ConsistOf("a", "b"))
	})
})
