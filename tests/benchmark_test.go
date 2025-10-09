package tests

import (
	"testing"

	"delivery-management/internal/models"
	"delivery-management/internal/utils"
)

func BenchmarkPasswordHashing(b *testing.B) {
	password := "testpassword123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := utils.HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInputSanitization(b *testing.B) {
	input := "<script>alert('test')</script>Some normal text with special chars: &<>\"'"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.SanitizeInput(input)
	}
}

func BenchmarkStatusValidation(b *testing.B) {
	status := models.StatusCreated
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		status.IsValid()
		status.CanTransitionTo(models.StatusDispatched)
	}
}

func BenchmarkOrderResponseConversion(b *testing.B) {
	order := &models.Order{
		ID:          1,
		CustomerID:  1,
		Status:      models.StatusCreated,
		Items:       "Test items",
		Description: "Test description",
		Address:     "Test address",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order.ToResponse()
	}
}