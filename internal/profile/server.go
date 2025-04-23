package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/darvenommm/dating-bot-service/internal/filter"
	"github.com/darvenommm/dating-bot-service/internal/orm"
	commonv1 "github.com/darvenommm/dating-bot-service/pkg/api/common/v1"
	profilev1 "github.com/darvenommm/dating-bot-service/pkg/api/profile/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/gorm"
)

type ProfileServer struct {
	profilev1.UnimplementedProfileServiceServer
	orm   *orm.ORM
	redis *redis.Client
}

func NewServer(orm *orm.ORM, redis *redis.Client) *ProfileServer {
	return &ProfileServer{orm: orm, redis: redis}
}

func (s *ProfileServer) SetProfile(
	_ context.Context,
	request *profilev1.SetProfileRequest,
) (*profilev1.SetProfileResponse, error) {
	db := s.orm.DB()

	profile := Profile{UserID: int(request.GetUserId())}
	attrs := Profile{
		FullName:    request.GetFullName(),
		Gender:      Gender(request.GetGender()),
		Age:         uint(request.GetAge()),
		Description: request.GetDescription(),
		Photo:       request.GetPhoto(),
	}

	tx := db.Where("user_id = ?", request.GetUserId()).Assign(attrs).FirstOrCreate(&profile)
	if err := tx.Error; err != nil {
		return nil, status.Errorf(codes.Internal, "couldn't set profile: %v", err)
	}

	return &profilev1.SetProfileResponse{}, nil
}

func (s *ProfileServer) GetProfile(
	_ context.Context,
	req *profilev1.GetProfileRequest,
) (*profilev1.GetProfileResponse, error) {
	var p Profile
	db := s.orm.DB()

	err := db.Where("user_id = ?", req.GetUserId()).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.NotFound, "not found profile user_id=%d", req.GetUserId())
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "db fetch profile: %v", err)
	}

	var desc *wrapperspb.StringValue
	if p.Description != "" {
		desc = wrapperspb.String(p.Description)
	}

	var photo *wrapperspb.BytesValue
	if len(p.Photo) > 0 {
		photo = wrapperspb.Bytes(p.Photo)
	}

	return &profilev1.GetProfileResponse{
		Profile: &profilev1.Profile{
			FullName:    p.FullName,
			Gender:      commonv1.Gender(p.Gender),
			Age:         uint32(p.Age),
			Description: desc,
			Photo:       photo,
		},
	}, nil
}

func (s *ProfileServer) GetRecommendation(
	ctx context.Context,
	req *profilev1.GetRecommendationRequest,
) (*profilev1.GetRecommendationResponse, error) {
	uid := req.GetUserId()
	key := fmt.Sprintf("recommendations:%d", uid)

	raw, err := s.redis.LPop(ctx, key).Result()
	if err == nil {
		var prof profilev1.Profile
		if err := json.Unmarshal([]byte(raw), &prof); err != nil {
			return nil, status.Errorf(codes.Internal, "bad cached profile: %v", err)
		}

		return &profilev1.GetRecommendationResponse{Profile: &prof}, nil
	}
	if err != redis.Nil {
		return nil, status.Errorf(codes.Internal, "redis LPop: %v", err)
	}

	var f filter.Filter
	if ferr := s.orm.DB().
		Where("user_id = ?", uid).
		First(&f).Error; errors.Is(ferr, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.NotFound, "filter not found for %d", uid)
	} else if ferr != nil {
		return nil, status.Errorf(codes.Internal, "db filter: %v", ferr)
	}

	var cand []Profile
	if dberr := s.orm.DB().
		Where("gender = ? AND age BETWEEN ? AND ?", f.Gender, f.MinAge, f.MaxAge).
		Where("user_id <> ?", uid).
		Where(`
				NOT EXISTS (
					SELECT 1 FROM user_actions as ua
					WHERE (ua.from_user_id = ? AND ua.to_user_id = profiles.user_id)
				)
			`, uid).
		Find(&cand).Error; dberr != nil {
		return nil, status.Errorf(codes.Internal, "db candidates: %v", dberr)
	}

	if len(cand) == 0 {
		return nil, status.Errorf(codes.NotFound, "no recommendations for %d", uid)
	}

	for _, p := range cand {
		var desc *wrapperspb.StringValue
		if p.Description != "" {
			desc = wrapperspb.String(p.Description)
		}
		var photo *wrapperspb.BytesValue
		if len(p.Photo) > 0 {
			photo = wrapperspb.Bytes(p.Photo)
		}
		prof := profilev1.Profile{
			FullName:    p.FullName,
			Gender:      commonv1.Gender(p.Gender),
			Age:         uint32(p.Age),
			Description: desc,
			Photo:       photo,
		}
		b, _ := json.Marshal(&prof)
		if err := s.redis.RPush(ctx, key, b).Err(); err != nil {
			log.Printf("redis RPush error: %v", err)
		}
	}

	raw, err = s.redis.LPop(ctx, key).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "redis LPop after fill: %v", err)
	}
	var first profilev1.Profile
	if err := json.Unmarshal([]byte(raw), &first); err != nil {
		return nil, status.Errorf(codes.Internal, "bad profile after fill: %v", err)
	}

	return &profilev1.GetRecommendationResponse{Profile: &first}, nil
}
