-- ledger_account_type_config 預設錨定科目
-- BANK_ACCOUNT → 1101 現金及約當現金
-- LOAN         → 2201 長期借款
-- CREDIT_CARD  → 2101 應付款項（含 2101-01 信用卡應付款）
INSERT INTO ledger_account_type_config(merchant_id, type, account_id)
VALUES
    (1, 'BANK_ACCOUNT', '1101'),
    (1, 'LOAN',         '2201'),
    (1, 'CREDIT_CARD',  '2101')
ON CONFLICT(merchant_id, type) DO NOTHING;

-- asset_type_account_config 預設投資科目
-- unrealized_loss(5402-21) 與 oci(3102-02) 四種 AssetType 初始共用相同值
INSERT INTO asset_type_account_config(
    merchant_id, asset_type,
    realized_gain_account_id,
    realized_loss_account_id,
    unrealized_gain_account_id,
    unrealized_loss_account_id,
    oci_account_id,
    fee_account_id,
    tax_account_id,
    account_id
) VALUES
    (1, 'STOCK', '4203-01', '5402-11', '4204-01', '5402-21', '3102-02', '5402-01', '5402-05', '1203-02'),
    (1, 'FUND',  '4203-02', '5402-12', '4204-02', '5402-21', '3102-02', '5402-02', '5402-06', '1203-04'),
    (1, 'GOLD',  '4203-03', '5402-13', '4204-03', '5402-21', '3102-02', '5402-03', '5402-07', '1203'),
    (1, 'FX',    '4203-04', '5402-14', '4204-04', '5402-21', '3102-02', '5402-04', '5402-08', '1203')
ON CONFLICT(merchant_id, asset_type) DO UPDATE SET
    account_id = COALESCE(asset_type_account_config.account_id, excluded.account_id);
