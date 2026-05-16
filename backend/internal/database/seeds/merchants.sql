INSERT INTO merchants (merchant_id, name, display_name)
VALUES
    (1, '測試商戶', '測試商戶')
ON CONFLICT (merchant_id) DO NOTHING;

INSERT INTO user_merchants (user_id, merchant_id)
VALUES
    (1, 1)
    ON CONFLICT (user_id, merchant_id) DO NOTHING;