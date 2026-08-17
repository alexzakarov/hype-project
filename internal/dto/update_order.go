package dto

import "main/internal/order/domain/models"

type UpdateShoppingItemsReqDto struct {
	ShopItems []*models.ShopItem `json:"shopItems" bson:"shopItems,omitempty" validate:"required"`
}
