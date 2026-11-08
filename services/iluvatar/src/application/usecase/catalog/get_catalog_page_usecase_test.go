package catalog

import (
	"errors"
	"iluvatar/src/domain/models"
	"iluvatar/src/shared/testdoubles"
	"reflect"
	"testing"
)

func TestGetCatalogPage(t *testing.T) {
	services := []*models.Service{{ID: "1", Name: "PAE"}}
	categories := []*models.Category{{ID: "10", Name: "Salud"}}
	errServices := errors.New("services failed")
	errCategories := errors.New("categories failed")

	tests := []struct {
		name               string
		servicesErr        error
		categoriesErr      error
		want               *models.CatalogPage
		wantErr            error
		wantCategoriesCall bool
	}{
		{
			name:               "combines services and categories",
			want:               &models.CatalogPage{Services: services, Categories: categories},
			wantCategoriesCall: true,
		},
		{
			name:        "services failure is returned and categories are not queried",
			servicesErr: errServices,
			wantErr:     errServices,
		},
		{
			name:               "categories failure is returned",
			categoriesErr:      errCategories,
			wantErr:            errCategories,
			wantCategoriesCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoriesCalled := false
			serviceRepository := &testdoubles.ServiceRepository{
				GetServicesFunc: func() ([]*models.Service, error) {
					if tt.servicesErr != nil {
						return nil, tt.servicesErr
					}
					return services, nil
				},
			}
			categoryRepository := &testdoubles.CategoryRepository{
				GetCategoriesFunc: func() ([]*models.Category, error) {
					categoriesCalled = true
					if tt.categoriesErr != nil {
						return nil, tt.categoriesErr
					}
					return categories, nil
				},
			}

			got, err := NewCatalogUseCase(serviceRepository, categoryRepository).GetCatalogPage()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("page = %+v, want %+v", got, tt.want)
			}
			if categoriesCalled != tt.wantCategoriesCall {
				t.Errorf("categories queried = %v, want %v", categoriesCalled, tt.wantCategoriesCall)
			}
		})
	}
}
