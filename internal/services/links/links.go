package links

import (
	"errors"
	"time"
)

var (
	EmptyLinkErr = errors.New("email is empty")
)

type Repository interface {
	SaveLink(link string, linkTtl time.Duration, userId int64) error
}

type Service struct {
	linksRepository Repository
}

func New(linksRepository Repository) *Service {
	return &Service{
		linksRepository: linksRepository,
	}
}

func (s *Service) SaveLink(link string, linkTtl time.Duration, userId int64) error {
	if link == "" {
		return EmptyLinkErr
	}

	return s.linksRepository.SaveLink(link, linkTtl, userId)
}
