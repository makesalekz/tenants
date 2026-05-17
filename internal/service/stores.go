package service

import (
	"context"
	"time"

	v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/tenants/ent"
	"gitlab.calendaria.team/services/tenants/internal/biz"
	"gitlab.calendaria.team/services/tenants/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v2/auth"
)

type StoresService struct {
	v1.UnimplementedStoresServer

	su *biz.StoresUsecase
}

func NewStoresService(
	su *biz.StoresUsecase,
) *StoresService {
	return &StoresService{
		su: su,
	}
}

func (s *StoresService) CreateStore(ctx context.Context, req *v1.CreateStoreRequest) (*v1.StoreReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	dto := data.StoreDto{
		TenantID:  tenantID,
		Name:      req.GetName(),
		Address:   req.GetAddress(),
		Phone:     req.GetPhone(),
		WorkHours: req.GetWorkHours(),
	}
	if req.Lat != nil {
		lat := *req.Lat
		dto.Lat = &lat
	}
	if req.Lon != nil {
		lon := *req.Lon
		dto.Lon = &lon
	}

	store, err := s.su.CreateStore(ctx, dto)
	if err != nil {
		return nil, err
	}
	return &v1.StoreReply{
		Store: replyStore(store),
	}, nil
}

func (s *StoresService) UpdateStore(ctx context.Context, req *v1.UpdateStoreRequest) (*v1.StoreReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	dto := data.StoreDto{
		ID:        req.GetStoreId(),
		TenantID:  tenantID,
		Name:      req.GetName(),
		Address:   req.GetAddress(),
		Phone:     req.GetPhone(),
		WorkHours: req.GetWorkHours(),
		IsActive:  req.GetIsActive(),
	}
	if req.Lat != nil {
		lat := *req.Lat
		dto.Lat = &lat
	}
	if req.Lon != nil {
		lon := *req.Lon
		dto.Lon = &lon
	}

	store, err := s.su.UpdateStore(ctx, dto)
	if err != nil {
		return nil, err
	}
	return &v1.StoreReply{
		Store: replyStore(store),
	}, nil
}

func (s *StoresService) DeleteStore(ctx context.Context, req *v1.StoreRequest) (*utils_v1.EmptyReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	err := s.su.DeleteStore(ctx, req.GetStoreId(), tenantID)
	if err != nil {
		return nil, err
	}
	return &utils_v1.EmptyReply{}, nil
}

func (s *StoresService) GetStore(ctx context.Context, req *v1.StoreRequest) (*v1.StoreReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	store, err := s.su.GetStore(ctx, req.GetStoreId(), tenantID)
	if err != nil {
		return nil, err
	}
	return &v1.StoreReply{
		Store: replyStore(store),
	}, nil
}

func (s *StoresService) ListStores(ctx context.Context, req *v1.ListStoresRequest) (*v1.ListStoresReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	list, err := s.su.ListStores(ctx, data.StoresListFilter{
		TenantID:   tenantID,
		OnlyActive: req.GetOnlyActive(),
	}, req.GetPaginate())
	if err != nil {
		return nil, err
	}
	return &v1.ListStoresReply{
		Stores:   replyStores(list.Stores),
		Paginate: list.Paginate,
	}, nil
}

func (s *StoresService) SetStoreResponsible(ctx context.Context, req *v1.SetStoreResponsibleRequest) (*v1.StoreReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorEmptyActorId("empty tenant id")
	}

	store, err := s.su.SetResponsible(ctx, req.GetStoreId(), tenantID, req.GetMemberId())
	if err != nil {
		return nil, err
	}
	return &v1.StoreReply{
		Store: replyStore(store),
	}, nil
}

func replyStore(s *ent.Store) *v1.Store {
	result := v1.Store{
		Id:        s.ID,
		TenantId:  s.TenantID,
		Name:      s.Name,
		Address:   s.Address,
		Phone:     s.Phone,
		WorkHours: s.WorkHours,
		IsActive:  s.IsActive,
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
		UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
	}
	if s.Lat != nil {
		result.Lat = s.Lat
	}
	if s.Lon != nil {
		result.Lon = s.Lon
	}
	if s.ResponsibleID != nil {
		result.ResponsibleId = s.ResponsibleID
	}
	return &result
}

func replyStores(stores []*ent.Store) []*v1.Store {
	result := make([]*v1.Store, len(stores))
	for i, s := range stores {
		result[i] = replyStore(s)
	}
	return result
}
