package filter

import (
	"context"

	"github.com/darvenommm/dating-bot-service/internal/orm"
	filterv1 "github.com/darvenommm/dating-bot-service/pkg/api/filter/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FilterServer struct {
	filterv1.UnimplementedFilterServiceServer
	orm *orm.ORM
}

func NewServer(orm *orm.ORM) *FilterServer {
	return &FilterServer{orm: orm}
}

func (s *FilterServer) SetFilter(
	_ context.Context,
	request *filterv1.SetFilterRequest,
) (*filterv1.SetFilterResponse, error) {
	db := s.orm.DB()

	filter := Filter{UserID: int(request.GetUserId())}
	attrs := Filter{
		Gender: Gender(request.GetGender()),
		MinAge: uint(request.GetMinAge()),
		MaxAge: uint(request.GetMaxAge()),
	}

	tx := db.Where("user_id = ?", request.GetUserId()).Assign(attrs).FirstOrCreate(&filter)
	if err := tx.Error; err != nil {
		return nil, status.Errorf(codes.Internal, "couldn't set filter: %v", err)
	}

	return &filterv1.SetFilterResponse{}, nil
}
