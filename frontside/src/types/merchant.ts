export interface Merchant {
    merchant_id: number;
    name:   string;
    display_name: string;
    currency: string;
    status: string;
    created_at: string;
}

export interface SelecedtMerchant {
    merchant_id: number;
}

export interface UpdateMerchantRequest {
    name: string;
    display_name: string;
    currency: string;
}