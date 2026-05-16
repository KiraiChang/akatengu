INSERT INTO sys_accounts(merchant_id, sys_code, description, account_id)
VALUES
    (1,"SYS:ASSET:PREPAID_INTEREST", "預付利息（資產）", "1104-04"),
    (1,"SYS:EXPENSE:LOAN:INTEREST", "貸款利息費用", "5401-01"),
    (1,"SYS:EXPENSE:CREDIT_CARD:INTEREST", "信用卡利息費用", "5401-02"),
    (1,"SYS:EQUITY:CLOSE_NET_INCOME", "結清淨收益（權益）", "3102-01"),
    (1,"SYS:EQUITY:EQUITY_OPENING", "期初權益", "3101-01"),

    (1,"SYS:INCOME:FVTPL_STOCK_NET_INCOME", "按公允價值衡量之股票投資收益","4203-01"),
    (1,"SYS:INCOME:FVTPL_FUND_NET_INCOME", "按公允價值衡量之基金投資收益","4203-02"),
    (1,"SYS:INCOME:FVTPL_GOLD_NET_INCOME", "按公允價值衡量之貴金屬投資收益","4203-03"),
    (1,"SYS:INCOME:FVTPL_FX_NET_INCOME", "按公允價值衡量之外幣投資收益","4203-04"),

    (1,"SYS:Expense:FVTPL_STOCK_LOSS", "按公允價值衡量之股票投資損失","5402-11"),
    (1,"SYS:Expense:FVTPL_FUND_LOSS", "按公允價值衡量之基金投資損失","5402-12"),
    (1,"SYS:Expense:FVTPL_GOLD_LOSS", "按公允價值衡量之貴金屬投資損失","5402-13"),
    (1,"SYS:Expense:FVTPL_FX_LOSS", "按公允價值衡量之外幣投資損失","5402-14"),

    (1,"SYS:EXPENSE:INVESTMENT_STOCK_FEE", "投資股票手續費","5402-01"),
    (1,"SYS:EXPENSE:INVESTMENT_FUND_FEE", "投資基金手續費","5402-02"),
    (1,"SYS:EXPENSE:INVESTMENT_GOLD_FEE", "投資貴金屬手續費","5402-03"),
    (1,"SYS:EXPENSE:INVESTMENT_FX_FEE", "投資外匯手續費","5402-04"),
    (1,"SYS:EXPENSE:INVESTMENT_STOCK_TAX", "投資股票稅金","5402-05"),
    (1,"SYS:EXPENSE:INVESTMENT_FUND_TAX", "投資基金稅金","5402-06"),
    (1,"SYS:EXPENSE:INVESTMENT_GOLD_TAX", "投資貴金屬稅金","5402-07"),
    (1,"SYS:EXPENSE:INVESTMENT_FX_TAX", "投資外幣稅金","5402-08"),

    (1,"SYS:INCOME:FVTPL_STOCK_UNREALIZED", "FVTPL 股票未實現評價利益","4204-01"),
    (1,"SYS:INCOME:FVTPL_FUND_UNREALIZED",  "FVTPL 基金未實現評價利益","4204-02"),
    (1,"SYS:INCOME:FVTPL_GOLD_UNREALIZED",  "FVTPL 黃金未實現評價利益","4204-03"),
    (1,"SYS:INCOME:FVTPL_FX_UNREALIZED",    "FVTPL 外幣未實現評價利益","4204-04"),
    (1,"SYS:EXPENSE:FVTPL_UNREALIZED_LOSS", "FVTPL 公允價值變動損失","5402-21"),
    (1,"SYS:EQUITY:OCI",                    "其他綜合損益 FVOCI 未實現損益","3102-02")
    ON CONFLICT (merchant_id, sys_code) DO NOTHING;