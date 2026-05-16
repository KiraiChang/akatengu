import {apiFetch} from "./http";
import type {Merchant} from "../types/merchant";

interface SelectResponse {
    token: string;
}

export const select = async (merchant_id: number, token: string | null): Promise<SelectResponse> => {
    const response = await apiFetch('/api/merchant/select', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({merchant_id}),
    }, token);
    if (!response.ok) {
        const problem = await response.json();
        throw new Error(problem.detail ?? problem.title ?? '選擇商戶失敗');
    }
    return response.json();
}

export const list = async (token: string | null): Promise<Merchant[]> => {
    const response = await apiFetch('/api/merchant/list', {}, token);
    if (!response.ok) {
        const problem = await response.json();
        throw new Error(problem.detail ?? problem.title ?? '取得商戶列表失敗');
    }
    return response.json();
}

