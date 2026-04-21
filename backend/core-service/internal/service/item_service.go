package service

import (
	"log"

	"core-service/internal/dto"
	"core-service/internal/repository"
)

type ItemService struct {
	repo *repository.ItemRepository
}

func NewItemService(repo *repository.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) GetAllItems() []dto.Item {
	return s.repo.GetAll()
}

func (s *ItemService) CreateItem(req dto.CreateItemRequest) dto.Item {
	log.Printf("Creating item: %s", req.Name)
	return s.repo.Create(req.Name, req.Description)
}

func (s *ItemService) GetItem(id string) (dto.Item, bool) {
	return s.repo.GetByID(id)
}

func (s *ItemService) DeleteItem(id string) bool {
	return s.repo.Delete(id)
}

func (s *ItemService) UpdateItem(id string, req dto.CreateItemRequest) (dto.Item, bool) {
	return s.repo.Update(id, req.Name, req.Description)
}
