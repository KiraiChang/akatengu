INSERT INTO projection_checkpoints (projection_name, last_event_id) VALUES
                                                                        ('ACCOUNT',        0),
                                                                        ('TRANSACTION',    0)
ON CONFLICT (projection_name) DO NOTHING;

INSERT INTO accounts
(account_id, parent_id, name, type, normal_balance, is_summary, version)
VALUES

-- ============================================================
-- 1 資產 ASSET
-- ============================================================

-- 父科目
('110', NULL,  '流動資產',     'ASSET', 'DEBIT', TRUE,  1),
('120', NULL,  '非流動資產',   'ASSET', 'DEBIT', TRUE,  1),

-- 子科目 — 110 流動資產
('1101', '110', '現金及約當現金', 'ASSET', 'DEBIT', TRUE, 1),
('1102', '110', '短期投資',       'ASSET', 'DEBIT', TRUE, 1),
('1103', '110', '應收款項',       'ASSET', 'DEBIT', TRUE, 1),
('1104', '110', '預付款項',       'ASSET', 'DEBIT', TRUE, 1),

-- 孫科目 — 1101 現金及約當現金
('1101-01', '1101', '手頭現金',                    'ASSET', 'DEBIT', FALSE, 1),
('1101-02', '1101', '銀行活期存款',                'ASSET', 'DEBIT', FALSE, 1),
('1101-03', '1101', '銀行定期存款（3個月內到期）', 'ASSET', 'DEBIT', FALSE, 1),
('1101-04', '1101', '電子支付餘額',                'ASSET', 'DEBIT', FALSE, 1),

-- 孫科目 — 1102 短期投資
('1102-01', '1102', '透過損益按公允價值衡量之投資 (FVTPL)', 'ASSET', 'DEBIT', FALSE, 1),
('1102-02', '1102', '透過其他綜合損益按公允價值衡量之投資 (FVOCI)', 'ASSET', 'DEBIT', FALSE, 1),
('1102-03', '1102', '貨幣市場基金',                            'ASSET', 'DEBIT', FALSE, 1),

-- 孫科目 — 1103 應收款項
('1103-01', '1103', '應收薪資款',   'ASSET', 'DEBIT', FALSE, 1),
('1103-02', '1103', '應收租金',     'ASSET', 'DEBIT', FALSE, 1),
('1103-03', '1103', '應收利息',     'ASSET', 'DEBIT', FALSE, 1),
('1103-04', '1103', '應收股利',     'ASSET', 'DEBIT', FALSE, 1),
('1103-09', '1103', '其他應收款',   'ASSET', 'DEBIT', FALSE, 1),

-- 孫科目 — 1104 預付款項
('1104-01', '1104', '預付保險費',   'ASSET', 'DEBIT', FALSE, 1),
('1104-02', '1104', '預付房租',     'ASSET', 'DEBIT', FALSE, 1),
('1104-03', '1104', '預付訂閱費',   'ASSET', 'DEBIT', FALSE, 1),
('1104-04', '1104', '預付利息',   'ASSET', 'DEBIT', FALSE, 1),
('1104-09', '1104', '其他預付款',   'ASSET', 'DEBIT', FALSE, 1),

-- 子科目 — 120 非流動資產
('1201', '120', '不動產、廠房及設備 (IAS 16)', 'ASSET', 'DEBIT', TRUE, 1),
('1202', '120', '使用權資產 (IFRS 16)',         'ASSET', 'DEBIT', TRUE, 1),
('1203', '120', '長期投資',                     'ASSET', 'DEBIT', TRUE, 1),
('1204', '120', '無形資產及其他',               'ASSET', 'DEBIT', TRUE, 1),

-- 孫科目 — 1201 不動產、廠房及設備
('1201-01', '1201', '土地',                   'ASSET', 'DEBIT', FALSE, 1),
('1201-02', '1201', '房屋建築',               'ASSET', 'DEBIT', FALSE, 1),
('1201-03', '1201', '車輛',                   'ASSET', 'DEBIT', FALSE, 1),
('1201-04', '1201', '電腦及設備',             'ASSET', 'DEBIT', FALSE, 1),
('1201-09', '1201', '其他設備',               'ASSET', 'DEBIT', FALSE, 1),
('1201-99', '1201', '累計折舊（抵銷）',       'ASSET', 'CREDIT', FALSE, 1),

-- 孫科目 — 1202 使用權資產
('1202-01', '1202', '租賃住宅使用權資產',     'ASSET', 'DEBIT', FALSE, 1),
('1202-02', '1202', '租賃車輛使用權資產',     'ASSET', 'DEBIT', FALSE, 1),
('1202-99', '1202', '累計攤銷（抵銷）',       'ASSET', 'CREDIT', FALSE, 1),

-- 孫科目 — 1203 長期投資
('1203-01', '1203', '按攤銷後成本衡量之債券 (AC)', 'ASSET', 'DEBIT', FALSE, 1),
('1203-02', '1203', 'FVOCI 股票投資',              'ASSET', 'DEBIT', FALSE, 1),
('1203-03', '1203', '私募股權及創投',              'ASSET', 'DEBIT', FALSE, 1),
('1203-04', '1203', 'ETF / 指數型基金',            'ASSET', 'DEBIT', FALSE, 1),
('1203-05', '1203', '退休金帳戶（勞退自提）',      'ASSET', 'DEBIT', FALSE, 1),

-- 孫科目 — 1204 無形資產及其他
('1204-01', '1204', '軟體授權',         'ASSET', 'DEBIT', FALSE, 1),
('1204-02', '1204', '押金及保證金',     'ASSET', 'DEBIT', FALSE, 1),
('1204-03', '1204', '遞延所得稅資產',   'ASSET', 'DEBIT', FALSE, 1),

-- ============================================================
-- 2 負債 LIABILITY
-- ============================================================

-- 父科目
('210', NULL, '流動負債',   'LIABILITY', 'CREDIT', TRUE, 1),
('220', NULL, '非流動負債', 'LIABILITY', 'CREDIT', TRUE, 1),

-- 子科目 — 210 流動負債
('2101', '210', '應付款項',   'LIABILITY', 'CREDIT', TRUE, 1),
('2102', '210', '短期借款',   'LIABILITY', 'CREDIT', TRUE, 1),
('2103', '210', '預收款項',   'LIABILITY', 'CREDIT', TRUE, 1),

-- 孫科目 — 2101 應付款項
('2101-01', '2101', '信用卡應付款',         'LIABILITY', 'CREDIT', FALSE, 1),
('2101-02', '2101', '應付帳單（水電瓦斯）', 'LIABILITY', 'CREDIT', FALSE, 1),
('2101-03', '2101', '應付稅款',             'LIABILITY', 'CREDIT', FALSE, 1),
('2101-09', '2101', '其他應付款',           'LIABILITY', 'CREDIT', FALSE, 1),

-- 孫科目 — 2102 短期借款
('2102-01', '2102', '銀行信貸（1年內到期）', 'LIABILITY', 'CREDIT', FALSE, 1),
('2102-02', '2102', '親友借款（短期）',       'LIABILITY', 'CREDIT', FALSE, 1),

-- 孫科目 — 2103 預收款項
('2103-01', '2103', '預收租金收入', 'LIABILITY', 'CREDIT', FALSE, 1),
('2103-09', '2103', '其他預收款',   'LIABILITY', 'CREDIT', FALSE, 1),

-- 子科目 — 220 非流動負債
('2201', '220', '長期借款',           'LIABILITY', 'CREDIT', TRUE, 1),
('2202', '220', '租賃負債 (IFRS 16)', 'LIABILITY', 'CREDIT', TRUE, 1),
('2203', '220', '遞延稅負及其他',     'LIABILITY', 'CREDIT', TRUE, 1),

-- 孫科目 — 2201 長期借款
('2201-01', '2201', '房屋貸款',       'LIABILITY', 'CREDIT', FALSE, 1),
('2201-02', '2201', '汽車貸款',       'LIABILITY', 'CREDIT', FALSE, 1),
('2201-03', '2201', '學貸',           'LIABILITY', 'CREDIT', FALSE, 1),
('2201-09', '2201', '其他長期借款',   'LIABILITY', 'CREDIT', FALSE, 1),

-- 孫科目 — 2202 租賃負債
('2202-01', '2202', '租賃負債—流動部分',     'LIABILITY', 'CREDIT', FALSE, 1),
('2202-02', '2202', '租賃負債—非流動部分',   'LIABILITY', 'CREDIT', FALSE, 1),

-- 孫科目 — 2203 遞延稅負及其他
('2203-01', '2203', '遞延所得稅負債', 'LIABILITY', 'CREDIT', FALSE, 1),
('2203-09', '2203', '其他長期負債',   'LIABILITY', 'CREDIT', FALSE, 1),

-- ============================================================
-- 3 權益 EQUITY
-- ============================================================

-- 父科目
('310', NULL, '個人淨資產', 'EQUITY', 'CREDIT', TRUE, 1),

-- 子科目 — 310 個人淨資產
('3101', '310', '期初淨資產',     'EQUITY', 'CREDIT', TRUE, 1),
('3102', '310', '本期綜合損益',   'EQUITY', 'CREDIT', TRUE, 1),
('3103', '310', '提撥與提領',     'EQUITY', 'CREDIT', TRUE, 1),

-- 孫科目 — 3101
('3101-01', '3101', '期初結餘', 'EQUITY', 'CREDIT', FALSE, 1),

-- 孫科目 — 3102
('3102-01', '3102', '本期損益（由損益科目結轉）',          'EQUITY', 'CREDIT', TRUE, 1),
('3102-02', '3102', '其他綜合損益 (OCI)—FVOCI 未實現損益', 'EQUITY', 'CREDIT', FALSE, 1),
('3102-03', '3102', '其他綜合損益—匯兌差異',              'EQUITY', 'CREDIT', FALSE, 1),

('3102-11', '3102-01', '本期薪資及自僱所得結轉',              'EQUITY', 'CREDIT', FALSE, 1),
('3102-12', '3102-01', '本期投資損益結轉',              'EQUITY', 'CREDIT', FALSE, 1),
('3102-13', '3102-01', '本期其他收入結轉',              'EQUITY', 'CREDIT', FALSE, 1),

('3102-14', '3102', '本期費用結轉',              'EQUITY', 'DEBIT', FALSE, 1),

-- 孫科目 — 3103
('3103-01', '3103', '個人提領（生活支出轉入）', 'EQUITY', 'DEBIT',  FALSE, 1),
('3103-02', '3103', '個人增資（外部資金注入）', 'EQUITY', 'CREDIT', FALSE, 1),

-- ============================================================
-- 4 收入 INCOME
-- ============================================================

-- 父科目
('410', NULL, '勞動所得', 'INCOME', 'CREDIT', TRUE, 1),
('420', NULL, '投資所得', 'INCOME', 'CREDIT', TRUE, 1),
('430', NULL, '其他收入', 'INCOME', 'CREDIT', TRUE, 1),

-- 子科目 — 410
('4101', '410', '薪資所得',       'INCOME', 'CREDIT', TRUE, 1),
('4102', '410', '自僱及接案收入', 'INCOME', 'CREDIT', TRUE, 1),

-- 孫科目 — 4101
('4101-01', '4101', '月薪',       'INCOME', 'CREDIT', FALSE, 1),
('4101-02', '4101', '獎金及年終', 'INCOME', 'CREDIT', FALSE, 1),
('4101-03', '4101', '加班費',     'INCOME', 'CREDIT', FALSE, 1),

-- 孫科目 — 4102
('4102-01', '4102', '顧問費',         'INCOME', 'CREDIT', FALSE, 1),
('4102-02', '4102', '版稅及授權金',   'INCOME', 'CREDIT', FALSE, 1),
('4102-09', '4102', '其他自僱收入',   'INCOME', 'CREDIT', FALSE, 1),

-- 子科目 — 420
('4201', '420', '利息收入',         'INCOME', 'CREDIT', TRUE, 1),
('4202', '420', '股利收入',         'INCOME', 'CREDIT', TRUE, 1),
('4203', '420', '處分投資利得',     'INCOME', 'CREDIT', TRUE, 1),
('4204', '420', '公允價值變動利益 (FVTPL)', 'INCOME', 'CREDIT', TRUE, 1),

-- 孫科目 — 4201
('4201-01', '4201', '存款利息', 'INCOME', 'CREDIT', FALSE, 1),
('4201-02', '4201', '債券利息', 'INCOME', 'CREDIT', FALSE, 1),
('4201-03', '4201', 'P2P 利息', 'INCOME', 'CREDIT', FALSE, 1),

-- 孫科目 — 4202
('4202-01', '4202', '現金股利',                   'INCOME', 'CREDIT', FALSE, 1),
('4202-02', '4202', '股票股利（按公允價值認列）', 'INCOME', 'CREDIT', FALSE, 1),

-- 孫科目 — 4203
('4203-01', '4203', '股票處分利得',     'INCOME', 'CREDIT', FALSE, 1),
('4203-02', '4203', '基金處分利得',     'INCOME', 'CREDIT', FALSE, 1),
('4203-03', '4203', '不動產處分利得',   'INCOME', 'CREDIT', FALSE, 1),

-- 孫科目 — 4204
('4204-01', '4204', '股票未實現評價利益', 'INCOME', 'CREDIT', FALSE, 1),
('4204-02', '4204', '基金未實現評價利益', 'INCOME', 'CREDIT', FALSE, 1),

-- 子科目 — 430
('4301', '430', '租金收入', 'INCOME', 'CREDIT', TRUE, 1),
('4302', '430', '雜項收入', 'INCOME', 'CREDIT', TRUE, 1),

-- 孫科目 — 4301
('4301-01', '4301', '房屋租金收入',   'INCOME', 'CREDIT', FALSE, 1),
('4301-02', '4301', '設備出租收入',   'INCOME', 'CREDIT', FALSE, 1),

-- 孫科目 — 4302
('4302-01', '4302', '政府補助及退稅', 'INCOME', 'CREDIT', FALSE, 1),
('4302-02', '4302', '保險理賠收入',   'INCOME', 'CREDIT', FALSE, 1),
('4302-03', '4302', '拍賣及二手收入', 'INCOME', 'CREDIT', FALSE, 1),
('4302-09', '4302', '其他雜項收入',   'INCOME', 'CREDIT', FALSE, 1),

-- ============================================================
-- 5 費用 EXPENSE
-- ============================================================

-- 父科目
('510', NULL, '生活費用',     'EXPENSE', 'DEBIT', TRUE, 1),
('520', NULL, '保障費用',     'EXPENSE', 'DEBIT', TRUE, 1),
('530', NULL, '個人發展',     'EXPENSE', 'DEBIT', TRUE, 1),
('540', NULL, '投資相關費用', 'EXPENSE', 'DEBIT', TRUE, 1),
('550', NULL, '折舊攤銷',     'EXPENSE', 'DEBIT', TRUE, 1),

-- 子科目 — 510
('5101', '510', '居住費用', 'EXPENSE', 'DEBIT', TRUE, 1),
('5102', '510', '飲食費用', 'EXPENSE', 'DEBIT', TRUE, 1),
('5103', '510', '交通費用', 'EXPENSE', 'DEBIT', TRUE, 1),
('5104', '510', '醫療保健', 'EXPENSE', 'DEBIT', TRUE, 1),

-- 孫科目 — 5101
('5101-01', '5101', '房租費用（短期租賃或使用權資產攤銷）', 'EXPENSE', 'DEBIT', FALSE, 1),
('5101-02', '5101', '房屋貸款利息費用',                     'EXPENSE', 'DEBIT', FALSE, 1),
('5101-03', '5101', '物業管理費',                           'EXPENSE', 'DEBIT', FALSE, 1),
('5101-04', '5101', '水費',                                 'EXPENSE', 'DEBIT', FALSE, 1),
('5101-05', '5101', '電費',                                 'EXPENSE', 'DEBIT', FALSE, 1),
('5101-06', '5101', '瓦斯費',                               'EXPENSE', 'DEBIT', FALSE, 1),
('5101-07', '5101', '網路及電話費',                         'EXPENSE', 'DEBIT', FALSE, 1),

-- 孫科目 — 5102
('5102-01', '5102', '居家伙食',   'EXPENSE', 'DEBIT', FALSE, 1),
('5102-02', '5102', '外食及餐廳', 'EXPENSE', 'DEBIT', FALSE, 1),
('5102-03', '5102', '飲品及咖啡', 'EXPENSE', 'DEBIT', FALSE, 1),

-- 孫科目 — 5103
('5103-01', '5103', '大眾運輸',       'EXPENSE', 'DEBIT', FALSE, 1),
('5103-02', '5103', '計程車及共享車', 'EXPENSE', 'DEBIT', FALSE, 1),
('5103-03', '5103', '燃油費',         'EXPENSE', 'DEBIT', FALSE, 1),
('5103-04', '5103', '車輛保養維修',   'EXPENSE', 'DEBIT', FALSE, 1),
('5103-05', '5103', '停車費',         'EXPENSE', 'DEBIT', FALSE, 1),
('5103-06', '5103', '航空及長途交通', 'EXPENSE', 'DEBIT', FALSE, 1),

-- 孫科目 — 5104
('5104-01', '5104', '門診及住院醫療', 'EXPENSE', 'DEBIT', FALSE, 1),
('5104-02', '5104', '藥品及保健品',   'EXPENSE', 'DEBIT', FALSE, 1),
('5104-03', '5104', '健身及運動',     'EXPENSE', 'DEBIT', FALSE, 1),
('5104-04', '5104', '心理諮商',       'EXPENSE', 'DEBIT', FALSE, 1),

-- 子科目 — 520
('5201', '520', '保險費用', 'EXPENSE', 'DEBIT', TRUE, 1),
('5202', '520', '稅捐費用', 'EXPENSE', 'DEBIT', TRUE, 1),

-- 孫科目 — 5201
('5201-01', '5201', '人壽保險費',             'EXPENSE', 'DEBIT', FALSE, 1),
('5201-02', '5201', '醫療險費',               'EXPENSE', 'DEBIT', FALSE, 1),
('5201-03', '5201', '財產險費',               'EXPENSE', 'DEBIT', FALSE, 1),
('5201-04', '5201', '汽機車強制險及任意險',   'EXPENSE', 'DEBIT', FALSE, 1),

-- 孫科目 — 5202
('5202-01', '5202', '綜合所得稅',         'EXPENSE', 'DEBIT', FALSE, 1),
('5202-02', '5202', '房屋稅及地價稅',     'EXPENSE', 'DEBIT', FALSE, 1),
('5202-03', '5202', '汽車燃料稅及牌照稅', 'EXPENSE', 'DEBIT', FALSE, 1),
('5202-04', '5202', '健保補充保費',       'EXPENSE', 'DEBIT', FALSE, 1),
('5202-09', '5202', '其他稅捐',           'EXPENSE', 'DEBIT', FALSE, 1),

-- 子科目 — 530
('5301', '530', '教育費用',   'EXPENSE', 'DEBIT', TRUE, 1),
('5302', '530', '娛樂及社交', 'EXPENSE', 'DEBIT', TRUE, 1),

-- 孫科目 — 5301
('5301-01', '5301', '學費及學雜費',     'EXPENSE', 'DEBIT', FALSE, 1),
('5301-02', '5301', '書籍及學習材料',   'EXPENSE', 'DEBIT', FALSE, 1),
('5301-03', '5301', '線上課程及訂閱',   'EXPENSE', 'DEBIT', FALSE, 1),
('5301-04', '5301', '語言及證照考試費', 'EXPENSE', 'DEBIT', FALSE, 1),

-- 孫科目 — 5302
('5302-01', '5302', '影音串流訂閱',   'EXPENSE', 'DEBIT', FALSE, 1),
('5302-02', '5302', '旅遊及住宿',     'EXPENSE', 'DEBIT', FALSE, 1),
('5302-03', '5302', '禮品及節慶支出', 'EXPENSE', 'DEBIT', FALSE, 1),
('5302-04', '5302', '聚餐及社交',     'EXPENSE', 'DEBIT', FALSE, 1),

-- 子科目 — 540
('5401', '540', '金融費用',         'EXPENSE', 'DEBIT', TRUE, 1),
('5402', '540', '投資手續費及損失', 'EXPENSE', 'DEBIT', TRUE, 1),

-- 孫科目 — 5401
('5401-01', '5401', '貸款利息費用（非房貸）', 'EXPENSE', 'DEBIT', FALSE, 1),
('5401-02', '5401', '信用卡循環利息',         'EXPENSE', 'DEBIT', FALSE, 1),
('5401-03', '5401', '帳戶管理費',             'EXPENSE', 'DEBIT', FALSE, 1),
('5401-04', '5401', '信用卡分期手續費', 'EXPENSE', 'DEBIT', FALSE, 1),

-- 孫科目 — 5402
('5402-01', '5402', '股票交易手續費及證交稅', 'EXPENSE', 'DEBIT', FALSE, 1),
('5402-02', '5402', '基金申購/贖回手續費',    'EXPENSE', 'DEBIT', FALSE, 1),
('5402-03', '5402', '處分投資損失',           'EXPENSE', 'DEBIT', FALSE, 1),
('5402-04', '5402', '公允價值變動損失 (FVTPL)', 'EXPENSE', 'DEBIT', FALSE, 1),

-- 子科目 — 550
('5501', '550', '折舊費用', 'EXPENSE', 'DEBIT', TRUE, 1),
('5502', '550', '攤銷費用', 'EXPENSE', 'DEBIT', TRUE, 1),

-- 孫科目 — 5501
('5501-01', '5501', '房屋折舊',         'EXPENSE', 'DEBIT', FALSE, 1),
('5501-02', '5501', '車輛折舊',         'EXPENSE', 'DEBIT', FALSE, 1),
('5501-03', '5501', '電腦及設備折舊',   'EXPENSE', 'DEBIT', FALSE, 1),

-- 孫科目 — 5502
('5502-01', '5502', '使用權資產攤銷 (IFRS 16)', 'EXPENSE', 'DEBIT', FALSE, 1),
('5502-02', '5502', '無形資產攤銷',             'EXPENSE', 'DEBIT', FALSE, 1)
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO sys_accounts(sys_code, account_id)
VALUES
    ("SYS:ASSET:PREPAID_INTEREST", "1104-04"),
    ("SYS:EXPENSE:LOAN:INTEREST", "5401-01"),
    ("SYS:EXPENSE:CREDIT_CARD:INTEREST", "5401-02"),
    ("SYS:EQUITY:CLOSE_NET_INCOME", "3102-01"),
    ("SYS:EQUITY:EQUITY_OPENING", "3101-01")
ON CONFLICT (sys_code) DO NOTHING;

INSERT INTO users (username, password, status)
VALUES
    ("admin","$argon2id$v=19$m=65536,t=3,p=4$dAwHffvcmwhMyBdu17BV2A$zKcGOexREzx6GetSdqq6OwXmOHWP5cY1+MRyTV9iq/c","ACTIVE")
    ON CONFLICT (username) DO NOTHING;