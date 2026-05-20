INSERT INTO sys_accounts(merchant_id, sys_code, description, account_id)
VALUES
    (1,"SYS:ASSET:PREPAID_INTEREST", "預付利息（資產）", "1104-04"),
    (1,"SYS:EXPENSE:LOAN:INTEREST", "貸款利息費用", "5401-01"),
    (1,"SYS:EXPENSE:CREDIT_CARD:INTEREST", "信用卡利息費用", "5401-02"),
    (1,"SYS:EQUITY:CLOSE_NET_INCOME", "結清淨收益（權益）", "3102-01"),
    (1,"SYS:EQUITY:EQUITY_OPENING", "期初權益", "3101-01")
    ON CONFLICT (merchant_id, sys_code) DO NOTHING;