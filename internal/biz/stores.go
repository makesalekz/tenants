package biz

import (
	"context"

	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type StoresList struct {
	Stores   []*ent.Store
	Paginate *utils_v1.PaginateReply
}

// StoresUsecase handles store business logic.
type StoresUsecase struct {
	log  *log.Helper
	repo data.StoresRepo
}

// NewStoresUsecase creates a new StoresUsecase.
func NewStoresUsecase(
	logger log.Logger,
	repo data.StoresRepo,
) (*StoresUsecase, error) {
	return &StoresUsecase{
		log:  log.NewHelper(logger),
		repo: repo,
	}, nil
}

func (uc *StoresUsecase) CreateStore(ctx context.Context, dto data.StoreDto) (*ent.Store, error) {
	s, err := uc.repo.CreateStore(ctx, dto)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (uc *StoresUsecase) UpdateStore(ctx context.Context, dto data.StoreDto) (*ent.Store, error) {
	s, err := uc.repo.UpdateStore(ctx, dto)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, tenants_v1.ErrorNotFound("store not found")
		}
		return nil, err
	}
	return s, nil
}

func (uc *StoresUsecase) DeleteStore(ctx context.Context, id int64, tenantID int64) error {
	err := uc.repo.DeleteStore(ctx, id, tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			return tenants_v1.ErrorNotFound("store not found")
		}
		return err
	}
	return nil
}

func (uc *StoresUsecase) GetStore(ctx context.Context, id int64, tenantID int64) (*ent.Store, error) {
	s, err := uc.repo.GetStore(ctx, id, tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, tenants_v1.ErrorNotFound("store not found")
		}
		return nil, err
	}
	return s, nil
}

func (uc *StoresUsecase) ListStores(
	ctx context.Context,
	filter data.StoresListFilter,
	paginate *utils_v1.PaginateRequest,
) (*StoresList, error) {
	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	stores, err := uc.repo.ListStores(ctx, filter, paginate)
	if err != nil {
		return nil, err
	}

	total, err := uc.repo.CountListStores(ctx, filter)
	if err != nil {
		return nil, err
	}

	paginateReply := utils_v1.PaginateReply{
		Total: &total,
	}

	if len(stores) == int(paginate.GetLimit()) {
		paginateReply.FromId = &stores[len(stores)-1].ID
	}

	return &StoresList{
		Stores:   stores,
		Paginate: &paginateReply,
	}, nil
}

func (uc *StoresUsecase) SetResponsible(ctx context.Context, storeID int64, tenantID int64, memberID int64) (*ent.Store, error) {
	s, err := uc.repo.SetResponsible(ctx, storeID, tenantID, memberID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, tenants_v1.ErrorNotFound("store or member not found")
		}
		return nil, err
	}
	return s, nil
}
