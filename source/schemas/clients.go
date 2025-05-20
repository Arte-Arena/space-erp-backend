package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Contact struct {
	Name                    string `json:"name,omitempty" bson:"name,omitempty"`
	Email                   string `json:"email,omitempty" bson:"email,omitempty"`
	PersonType              string `json:"person_type,omitempty" bson:"person_type,omitempty"`
	IdentityCard            string `json:"identity_card,omitempty" bson:"identity_card,omitempty"`
	CPF                     string `json:"cpf,omitempty" bson:"cpf,omitempty"`
	CellPhone               string `json:"cell_phone,omitempty" bson:"cell_phone,omitempty"`
	ZipCode                 string `json:"zip_code,omitempty" bson:"zip_code,omitempty"`
	Address                 string `json:"address,omitempty" bson:"address,omitempty"`
	Number                  string `json:"number,omitempty" bson:"number,omitempty"`
	Complement              string `json:"complement,omitempty" bson:"complement,omitempty"`
	Neighborhood            string `json:"neighborhood,omitempty" bson:"neighborhood,omitempty"`
	City                    string `json:"city,omitempty" bson:"city,omitempty"`
	State                   string `json:"state,omitempty" bson:"state,omitempty"`
	CompanyName             string `json:"company_name,omitempty" bson:"company_name,omitempty"`
	CNPJ                    string `json:"cnpj,omitempty" bson:"cnpj,omitempty"`
	StateRegistration       string `json:"state_registration,omitempty" bson:"state_registration,omitempty"`
	BillingZipCode          string `json:"billing_zip_code,omitempty" bson:"billing_zip_code,omitempty"`
	BillingAddress          string `json:"billing_address,omitempty" bson:"billing_address,omitempty"`
	BillingNumber           string `json:"billing_number,omitempty" bson:"billing_number,omitempty"`
	BillingComplement       string `json:"billing_complement,omitempty" bson:"billing_complement,omitempty"`
	BillingNeighborhood     string `json:"billing_neighborhood,omitempty" bson:"billing_neighborhood,omitempty"`
	BillingCity             string `json:"billing_city,omitempty" bson:"billing_city,omitempty"`
	BillingState            string `json:"billing_state,omitempty" bson:"billing_state,omitempty"`
	DifferentBillingAddress bool   `json:"different_billing_address,omitempty" bson:"different_billing_address,omitempty"`
}

type Client struct {
	ID        bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Contact   Contact       `json:"contact,omitempty" bson:"contact,omitempty"`
	BudgetIDs []int         `json:"budget_ids,omitempty" bson:"budget_ids,omitempty"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}
