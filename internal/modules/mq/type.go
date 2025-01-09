package mq

type AddressAssistanceReqUpdatedMsg struct {
	AddressAssistanceReqID string `json:"address_assistance_request_id"`
}

type AdminUpdatedMsg struct {
	AdminID string
}

type DriverTopupUpdatedMsg struct {
	DriverTopupID string `json:"driver_topup_id"`
}

type DriverWithdrawalUpdatedMsg struct {
	DriverWithdrawalID string `json:"driver_withdrawal_id"`
}

type MerchantWithdrawalUpdatedMsg struct {
	MerchantWithdrawalID string `json:"merchant_withdrawal_id"`
}

type CustomerUpdatedMsg struct {
	CustomerID string `json:"customer_id"`
}

type DriverUpdatedMsg struct {
	DriverID string `json:"driver_id"`
}

type MenuUpdatedMsg struct {
	MenuItemID string `json:"menu_item_id"`
	StoreID    string `json:"store_id"`
}

type MerchantUpdatedMsg struct {
	MerchantID string `json:"merchant_id"`
}

type StoreUpdatedMsg struct {
	StoreID string `json:"store_id"`
}

type OrderUpdatedMsg struct {
	OrderID string `json:"order_id"`
}
