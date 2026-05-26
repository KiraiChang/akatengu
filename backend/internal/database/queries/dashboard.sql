-- name: GetMonthlyTrend :many
SELECT month, income, expense
FROM v_monthly_income_expense
WHERE merchant_id = @merchant_id
  AND month >= @from_month
  AND month <= @to_month
ORDER BY month ASC;
