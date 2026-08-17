package mappers

import (
	"main/internal/dto"
	orderService "main/proto"
)

func SearchResponseToProto(protoSearch *dto.OrderSearchResponseDto) *orderService.SearchRes {
	orders := make([]*orderService.Order, 0, len(protoSearch.Orders))
	for _, order := range protoSearch.Orders {
		orders = append(orders, OrderResponseDtoToProto(order))
	}
	return &orderService.SearchRes{
		Pagination: PaginationToProto(protoSearch.Pagination),
		Orders:     orders,
	}
}
