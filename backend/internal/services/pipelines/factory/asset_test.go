package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/repos/query"
	"context"
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// 確保 state 套件被識別為使用（type assertion 在閉包內）
var _ *state.AssetPurchasedState
var _ *state.AssetPurchasedWithInstallmentState
var _ *state.PrepaidCreatedWithInstallmentState

type stubPrepaidCategoryRepo struct {
	category    *projection.PrepaidCategory
	categoryErr error
}

func (s *stubPrepaidCategoryRepo) GetPrepaidCategoryByUUID(_ context.Context, _ string) (*projection.PrepaidCategory, error) {
	return s.category, s.categoryErr
}
func (s *stubPrepaidCategoryRepo) GetAllPrepaidCategoriesByMerchant(_ context.Context) ([]projection.PrepaidCategory, error) {
	return nil, nil
}

var _ query.PrepaidCategoryQueryRepo = (*stubPrepaidCategoryRepo)(nil)

// ─────────────────────────────────────────
// Stubs
// ─────────────────────────────────────────

type stubAccountRepo struct {
	ledger    *projection.LedgerAccount
	ledgerErr error
}

func (s *stubAccountRepo) GetAccount(_ context.Context, _ string) (*projection.Account, error) {
	return nil, nil
}
func (s *stubAccountRepo) GetLedger(_ context.Context, _ int64) (*projection.LedgerAccount, error) {
	return nil, nil
}
func (s *stubAccountRepo) GetAllAccounts(_ context.Context) ([]projection.Account, error) {
	return nil, nil
}
func (s *stubAccountRepo) GetAccountPaged(_ context.Context, _ model.PaginationParams) ([]projection.Account, int64, error) {
	return nil, 0, nil
}
func (s *stubAccountRepo) GetChildrenAccount(_ context.Context, _ string) ([]projection.Account, error) {
	return nil, nil
}
func (s *stubAccountRepo) GetLedgerPaged(_ context.Context, _ model.PaginationParams) ([]projection.LedgerAccount, int64, error) {
	return nil, 0, nil
}
func (s *stubAccountRepo) GetAllLedgers(_ context.Context) ([]projection.LedgerAccount, error) {
	return nil, nil
}
func (s *stubAccountRepo) GetLedgerByUuid(_ context.Context, _ string) (*projection.LedgerAccount, error) {
	return s.ledger, s.ledgerErr
}

var _ query.AccountRepo = (*stubAccountRepo)(nil)

type stubSysRepo struct {
	codes    []db.SysAccount
	codesErr error
}

func (s *stubSysRepo) GetSysAccount(_ context.Context) ([]db.SysAccount, error) {
	return s.codes, s.codesErr
}

var _ query.SysRepo = (*stubSysRepo)(nil)

type stubFixedAssetRepo struct {
	asset    *projection.FixedAsset
	assetErr error
}

func (s *stubFixedAssetRepo) GetFixedAssetByID(_ context.Context, _ int64) (*projection.FixedAsset, error) {
	return nil, nil
}
func (s *stubFixedAssetRepo) GetFixedAssetByUUID(_ context.Context, _ string) (*projection.FixedAsset, error) {
	return s.asset, s.assetErr
}
func (s *stubFixedAssetRepo) GetActiveFixedAssetsByMerchant(_ context.Context) ([]projection.FixedAsset, error) {
	return nil, nil
}
func (s *stubFixedAssetRepo) GetAllFixedAssetsByMerchant(_ context.Context) ([]projection.FixedAsset, error) {
	return nil, nil
}
func (s *stubFixedAssetRepo) GetFixedAssetDepreciationsByAssetID(_ context.Context, _ int64) ([]projection.FixedAssetDepreciation, error) {
	return nil, nil
}

var _ query.FixedAssetQueryRepo = (*stubFixedAssetRepo)(nil)

type stubFixedAssetCategoryRepo struct {
	category    *projection.FixedAssetCategory
	categoryErr error
}

func (s *stubFixedAssetCategoryRepo) GetFixedAssetCategoryByUUID(_ context.Context, _ string) (*projection.FixedAssetCategory, error) {
	return s.category, s.categoryErr
}
func (s *stubFixedAssetCategoryRepo) GetAllFixedAssetCategoriesByMerchant(_ context.Context) ([]projection.FixedAssetCategory, error) {
	return nil, nil
}

var _ query.FixedAssetCategoryQueryRepo = (*stubFixedAssetCategoryRepo)(nil)

type stubPeriodRepo struct {
	period    *projection.PeriodClosing
	periodErr error
}

func (s *stubPeriodRepo) GetByPeriod(_ context.Context, _ enums.PeriodType, _ string) (*projection.PeriodClosing, error) {
	return s.period, s.periodErr
}
func (s *stubPeriodRepo) GetByID(_ context.Context, _ int64) (*projection.PeriodClosing, error) {
	return nil, nil
}
func (s *stubPeriodRepo) GetLatestClosed(_ context.Context, _ enums.PeriodType) (*projection.PeriodClosing, error) {
	return nil, nil
}
func (s *stubPeriodRepo) IsDateInClosedPeriod(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (s *stubPeriodRepo) AssertNoUnresolvedAdjustments(_ context.Context, _, _ string) error {
	return nil
}
func (s *stubPeriodRepo) GetPeriodPagedByType(_ context.Context, _ enums.PeriodType, _ model.PaginationParams) ([]projection.PeriodClosing, int64, error) {
	return nil, 0, nil
}

var _ query.PeriodRepo = (*stubPeriodRepo)(nil)

// ─────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────

func newStubQueryRepo(period *stubPeriodRepo, account *stubAccountRepo, fa *stubFixedAssetRepo) *query.Repo {
	return &query.Repo{
		Period:     period,
		Account:    account,
		FixedAsset: fa,
	}
}

func newStubQueryRepoWithSys(period *stubPeriodRepo, account *stubAccountRepo, fa *stubFixedAssetRepo, sys *stubSysRepo) *query.Repo {
	return &query.Repo{
		Period:             period,
		Account:            account,
		FixedAsset:         fa,
		Sys:                sys,
		FixedAssetCategory: &stubFixedAssetCategoryRepo{},
	}
}

func newStubQueryRepoWithSysAndCategory(period *stubPeriodRepo, account *stubAccountRepo, fa *stubFixedAssetRepo, sys *stubSysRepo, fac *stubFixedAssetCategoryRepo) *query.Repo {
	return &query.Repo{
		Period:             period,
		Account:            account,
		FixedAsset:         fa,
		Sys:                sys,
		FixedAssetCategory: fac,
	}
}

func sysCodesWithPrepaidInterest(sysCode string) []db.SysAccount {
	return []db.SysAccount{
		{SysCode: "SYS:ASSET:PREPAID_INTEREST", AccountId: sysCode},
	}
}

func testCtx() context.Context {
	return context.Background()
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("mustMarshal: %v", err))
	}
	return b
}

func appendCmd(p any) cmd.AppendCmd {
	return cmd.AppendCmd{
		AggregateID: "test-aggregate",
		Payload:     mustMarshal(p),
	}
}

// ─────────────────────────────────────────
// Spec runner
// ─────────────────────────────────────────

var _ = Describe("EventAssetPurchased Pipeline", func() {
	for _, s := range buildAssetPurchasedScenarios() {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			accountStub := &stubAccountRepo{ledger: s.ledger, ledgerErr: s.ledgerErr}
			periodStub := &stubPeriodRepo{period: s.period, periodErr: s.periodErr}
			faStub := &stubFixedAssetRepo{}
			facStub := &stubFixedAssetCategoryRepo{category: s.category, categoryErr: s.categoryErr}
			q := &query.Repo{
				Period:             periodStub,
				Account:            accountStub,
				FixedAsset:         faStub,
				FixedAssetCategory: facStub,
			}

			pipeline := NewEventAssetPurchasedPipeline(q)
			result, err := pipeline.Run(testCtx(), appendCmd(s.payload))

			if s.wantErrContain != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(s.wantErrContain))
			} else {
				Expect(err).NotTo(HaveOccurred())
				st, ok := result.State.(*state.AssetPurchasedState)
				Expect(ok).To(BeTrue())
				if s.checkState != nil {
					s.checkState(st)
				}
			}
		})
	}
})

var _ = Describe("EventAssetDepreciated Pipeline", func() {
	for _, s := range buildAssetDepreciatedScenarios() {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			accountStub := &stubAccountRepo{}
			periodStub := &stubPeriodRepo{period: s.period, periodErr: s.periodErr}
			faStub := &stubFixedAssetRepo{asset: s.asset, assetErr: s.assetErr}
			q := newStubQueryRepo(periodStub, accountStub, faStub)

			pipeline := NewEventAssetDepreciatedPipeline(q)
			result, err := pipeline.Run(testCtx(), appendCmd(s.payload))

			if s.wantErrContain != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(s.wantErrContain))
			} else {
				Expect(err).NotTo(HaveOccurred())
				st, ok := result.State.(*state.AssetDepreciatedState)
				Expect(ok).To(BeTrue())
				if s.checkState != nil {
					s.checkState(st)
				}
			}
		})
	}
})

var _ = Describe("EventAssetPurchasedWithInstallment Pipeline", func() {
	for _, s := range buildAssetPurchasedWithInstallmentScenarios() {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			accountStub := &stubAccountRepo{ledger: s.ledger, ledgerErr: s.ledgerErr}
			periodStub := &stubPeriodRepo{period: s.period, periodErr: s.periodErr}
			faStub := &stubFixedAssetRepo{}
			facStub := &stubFixedAssetCategoryRepo{category: s.category, categoryErr: s.categoryErr}
			sysStub := &stubSysRepo{codes: sysCodesWithPrepaidInterest(s.sysCode)}
			q := newStubQueryRepoWithSysAndCategory(periodStub, accountStub, faStub, sysStub, facStub)

			pipeline := NewEventAssetPurchasedWithInstallmentPipeline(q)
			result, err := pipeline.Run(testCtx(), appendCmd(s.payload))

			if s.wantErrContain != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(s.wantErrContain))
			} else {
				Expect(err).NotTo(HaveOccurred())
				st, ok := result.State.(*state.AssetPurchasedWithInstallmentState)
				Expect(ok).To(BeTrue())
				if s.checkState != nil {
					s.checkState(st)
				}
			}
		})
	}
})

var _ = Describe("EventPrepaidCreatedWithInstallment Pipeline", func() {
	for _, s := range buildPrepaidCreatedWithInstallmentScenarios() {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			accountStub := &stubAccountRepo{ledger: s.ledger, ledgerErr: s.ledgerErr}
			periodStub := &stubPeriodRepo{period: s.period, periodErr: s.periodErr}
			faStub := &stubFixedAssetRepo{}
			pacStub := &stubPrepaidCategoryRepo{category: s.category, categoryErr: s.categoryErr}
			sysStub := &stubSysRepo{codes: sysCodesWithPrepaidInterest(s.sysCode)}
			q := &query.Repo{
				Period:          periodStub,
				Account:         accountStub,
				FixedAsset:      faStub,
				PrepaidCategory: pacStub,
				Sys:             sysStub,
			}

			pipeline := NewEventPrepaidCreatedWithInstallmentPipeline(q)
			result, err := pipeline.Run(testCtx(), appendCmd(s.payload))

			if s.wantErrContain != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(s.wantErrContain))
			} else {
				Expect(err).NotTo(HaveOccurred())
				st, ok := result.State.(*state.PrepaidCreatedWithInstallmentState)
				Expect(ok).To(BeTrue())
				if s.checkState != nil {
					s.checkState(st)
				}
			}
		})
	}
})

var _ = Describe("EventAssetDisposed Pipeline", func() {
	for _, s := range buildAssetDisposedScenarios() {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			accountStub := &stubAccountRepo{ledger: s.proceedsLedger, ledgerErr: s.ledgerErr}
			periodStub := &stubPeriodRepo{}
			faStub := &stubFixedAssetRepo{asset: s.asset, assetErr: s.assetErr}
			q := newStubQueryRepo(periodStub, accountStub, faStub)

			pipeline := NewEventAssetDisposedPipeline(q)
			result, err := pipeline.Run(testCtx(), appendCmd(s.payload))

			if s.wantErrContain != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(s.wantErrContain))
			} else {
				Expect(err).NotTo(HaveOccurred())
				st, ok := result.State.(*state.AssetDisposedState)
				Expect(ok).To(BeTrue())
				if s.checkState != nil {
					s.checkState(st)
				}
			}
		})
	}
})
