package grpcServer

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kode/internal/entity"
	"kode/internal/transport/grpc/generated"
)

type serverAPI struct {
	medService entity.MedServiceInterface
	generated.UnimplementedMedicineServiceServer
}

func Register(s *grpc.Server, medService entity.MedServiceInterface) {
	generated.RegisterMedicineServiceServer(s, &serverAPI{medService: medService})
}

func (s *serverAPI) AddSchedule(ctx context.Context, req *generated.AddScheduleRequest) (*generated.AddScheduleResponse, error) {
	if req.GetName() == "" || req.GetUserId() < 0 || req.GetTakingDuration() < 0 || req.GetTreatmentDuration() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	id, err := s.medService.AddSchedule(ctx, req.GetName(), req.GetUserId(), req.GetTakingDuration(), req.GetTreatmentDuration())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &generated.AddScheduleResponse{Id: id}, nil
}

func (s *serverAPI) Schedules(ctx context.Context, req *generated.SchedulesRequest) (*generated.SchedulesResponse, error) {
	if req.GetUserId() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	ids, err := s.medService.Schedules(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &generated.SchedulesResponse{SchedulesId: ids}, nil
}

func (s *serverAPI) Schedule(ctx context.Context, req *generated.ScheduleRequest) (*generated.ScheduleResponse, error) {
	if req.GetUserId() < 0 || req.GetScheduleId() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	med, err := s.medService.Schedule(ctx, req.GetUserId(), req.GetScheduleId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &generated.ScheduleResponse{
		Id:                med.Id,
		Name:              med.Name,
		TakingDuration:    med.TakingDuration,
		TreatmentDuration: med.TreatmentDuration,
		UserId:            med.UserId,
		Schedule:          med.Schedule,
		Date:              med.Date.String(),
	}, nil
}

func (s *serverAPI) NextTakings(ctx context.Context, req *generated.NextTakingsRequest) (*generated.NextTakingsResponse, error) {
	if req.GetUserId() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	meds, err := s.medService.NextTakings(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &generated.NextTakingsResponse{Medicines: meds}, nil
}
