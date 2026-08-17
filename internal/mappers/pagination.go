package mappers

import (
	"main/internal/dto"
	orderService "main/proto"
)

func PaginationFromProto(protoPagination *orderService.Pagination) dto.Pagination {
	return dto.Pagination{
		TotalCount: protoPagination.GetTotalCount(),
		TotalPages: protoPagination.GetTotalPages(),
		Page:       protoPagination.GetPage(),
		Size:       protoPagination.GetSize(),
		HasMore:    protoPagination.GetHasMore(),
	}
}

func PaginationToProto(protoPagination dto.Pagination) *orderService.Pagination {
	return &orderService.Pagination{
		TotalCount: protoPagination.TotalCount,
		TotalPages: protoPagination.TotalPages,
		Page:       protoPagination.Page,
		Size:       protoPagination.Size,
		HasMore:    protoPagination.HasMore,
	}
}
