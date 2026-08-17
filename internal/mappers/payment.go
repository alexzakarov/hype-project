package mappers

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"main/internal/dto"
	"main/internal/order/domain/models"
	orderService "main/proto"
)

func PaymentFromProto(protoPayment *orderService.Payment) dto.Payment {
	return dto.Payment{
		PaymentID: protoPayment.GetID(),
		Timestamp: protoPayment.GetTimestamp().AsTime(),
	}
}

func PaymentResponseFromModel(payment models.Payment) dto.Payment {
	return dto.Payment{
		PaymentID: payment.PaymentID,
		Timestamp: payment.Timestamp,
	}
}

func PaymentToProto(payment dto.Payment) *orderService.Payment {
	return &orderService.Payment{
		ID:        payment.PaymentID,
		Timestamp: timestamppb.New(payment.Timestamp),
	}
}
