package sqlite

import "forum/models"

func (s *Storage) CreatePost(*models.Post) error {
	
	
	return nil
}

func (s *Storage) GetPost() error {
	return nil
}

func (s *Storage) GetPostByUserID() error {
	return nil
}

// Like system

func (s *Storage) AddLikePost() error {
	return nil
}

func (s *Storage) AddDisLikePost() error {
	return nil
}

func (s *Storage) DeleteLikeStatusPost() error {
	return nil
}
