package grpcServer

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kode/internal/storage"
	medicineProto "kode/proto/gen"
)

type medService interface {
	AddSchedule(ctx context.Context, name string, userId int64, takingDuration, treatmentDuration int32) (int64, error)
	Schedules(ctx context.Context, userId int64) ([]int64, error)
	Schedule(ctx context.Context, userId, scheduleId int64) (*storage.Medicine, error)
	NextTakings(ctx context.Context, userId int64) ([]*medicineProto.Medicines, error)
}

type serverAPI struct {
	medService medService
	medicineProto.UnimplementedMedicineServiceServer
}

func Register(s *grpc.Server, medService medService) {
	medicineProto.RegisterMedicineServiceServer(s, &serverAPI{medService: medService})
}

func (s *serverAPI) AddSchedule(ctx context.Context, req *medicineProto.AddScheduleRequest) (*medicineProto.AddScheduleResponse, error) {

	if req.GetName() == "" || req.GetUserId() < 0 || req.GetTakingDuration() < 0 || req.GetTreatmentDuration() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	id, err := s.medService.AddSchedule(ctx, req.GetName(), req.GetUserId(), req.GetTakingDuration(), req.GetTreatmentDuration())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &medicineProto.AddScheduleResponse{Id: id}, nil
}

func (s *serverAPI) Schedules(ctx context.Context, req *medicineProto.SchedulesRequest) (*medicineProto.SchedulesResponse, error) {
	if req.GetUserId() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	ids, err := s.medService.Schedules(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &medicineProto.SchedulesResponse{SchedulesId: ids}, nil
}

func (s *serverAPI) Schedule(ctx context.Context, req *medicineProto.ScheduleRequest) (*medicineProto.ScheduleResponse, error) {
	if req.GetUserId() < 0 || req.GetScheduleId() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	med, err := s.medService.Schedule(ctx, req.GetUserId(), req.GetScheduleId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &medicineProto.ScheduleResponse{
		Id:                med.Id,
		Name:              med.Name,
		TakingDuration:    med.TakingDuration,
		TreatmentDuration: med.TreatmentDuration,
		UserId:            med.UserId,
		Schedule:          med.Schedule,
		Date:              med.Date.String(),
	}, nil
}

func (s *serverAPI) NextTakings(ctx context.Context, req *medicineProto.NextTakingsRequest) (*medicineProto.NextTakingsResponse, error) {
	if req.GetUserId() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	meds, err := s.medService.NextTakings(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &medicineProto.NextTakingsResponse{Medicines: meds}, nil
}
