-- name: GetMonthlySummary :many
WITH monthly_totals AS (
    SELECT 
        i.currency,
        COALESCE(SUM(i.amount), 0) as total_income,
        COALESCE(SUM(e.amount), 0) as total_expenses,
        COALESCE(SUM(i.amount) - SUM(e.amount), 0) as total_savings
    FROM income i
    FULL OUTER JOIN expenses e ON 
        e.user_id = i.user_id 
        AND e.currency = i.currency
        AND DATE_TRUNC('month', e.date) = DATE_TRUNC('month', i.date)
        AND e.deleted_at IS NULL
    WHERE i.user_id = sqlc.arg(user_id)
        AND i.deleted_at IS NULL
        AND DATE_TRUNC('month', i.date) = DATE_TRUNC('month', sqlc.arg(date)::TIMESTAMPTZ)
    GROUP BY i.currency
),
top_categories AS (
    SELECT 
        e.category_id,
        c.name as category_name,
        e.currency,
        COUNT(*) as usage_count,
        SUM(e.amount) as total_spent,
        ROW_NUMBER() OVER (PARTITION BY e.currency ORDER BY SUM(e.amount) DESC) as rank
    FROM expenses e
    JOIN categories c ON e.category_id = c.id
    WHERE e.user_id = sqlc.arg(user_id)
        AND e.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND DATE_TRUNC('month', e.date) = DATE_TRUNC('month', sqlc.arg(date)::TIMESTAMPTZ)
    GROUP BY e.category_id, c.name, e.currency
)
SELECT 
    mt.*,
    json_agg(
        json_build_object(
            'category_id', tc.category_id,
            'category_name', tc.category_name,
            'usage_count', tc.usage_count,
            'total_spent', tc.total_spent
        )
    ) FILTER (WHERE tc.category_id IS NOT NULL) as top_categories
FROM monthly_totals mt
LEFT JOIN top_categories tc ON 
    tc.currency = mt.currency 
    AND tc.rank <= 5
GROUP BY 
    mt.currency, 
    mt.total_income, 
    mt.total_expenses, 
    mt.total_savings;

-- name: GetYearlySummary :many
WITH yearly_totals AS (
    SELECT 
        i.currency,
        COALESCE(SUM(i.amount), 0) as total_income,
        COALESCE(SUM(e.amount), 0) as total_expenses,
        COALESCE(SUM(i.amount) - SUM(e.amount), 0) as total_savings
    FROM income i
    FULL OUTER JOIN expenses e ON 
        e.user_id = i.user_id 
        AND e.currency = i.currency
        AND DATE_TRUNC('year', e.date) = DATE_TRUNC('year', i.date)
        AND e.deleted_at IS NULL
    WHERE i.user_id = sqlc.arg(user_id)
        AND i.deleted_at IS NULL
        AND DATE_TRUNC('year', i.date) = DATE_TRUNC('year', sqlc.arg(date)::TIMESTAMPTZ)
    GROUP BY i.currency
),
top_categories AS (
    SELECT 
        e.category_id,
        c.name as category_name,
        e.currency,
        COUNT(*) as usage_count,
        SUM(e.amount) as total_spent,
        ROW_NUMBER() OVER (PARTITION BY e.currency ORDER BY SUM(e.amount) DESC) as rank
    FROM expenses e
    JOIN categories c ON e.category_id = c.id
    WHERE e.user_id = sqlc.arg(user_id)
        AND e.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND DATE_TRUNC('year', e.date) = DATE_TRUNC('year', sqlc.arg(date)::TIMESTAMPTZ)
    GROUP BY e.category_id, c.name, e.currency
),
monthly_trend AS (
    SELECT 
        DATE_TRUNC('month', e.date) as month,
        e.currency,
        c.name as category_name,
        SUM(e.amount) as monthly_expenses
    FROM expenses e
    JOIN categories c ON e.category_id = c.id
    WHERE e.user_id = sqlc.arg(user_id)
        AND e.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND DATE_TRUNC('year', e.date) = DATE_TRUNC('year', sqlc.arg(date)::TIMESTAMPTZ)
    GROUP BY DATE_TRUNC('month', e.date), e.currency, c.name
    ORDER BY month
)
SELECT 
    yt.*,
    json_agg(
        json_build_object(
            'category_id', tc.category_id,
            'category_name', tc.category_name,
            'usage_count', tc.usage_count,
            'total_spent', tc.total_spent
        )
    ) FILTER (WHERE tc.category_id IS NOT NULL) as top_categories,
    json_agg(
        json_build_object(
            'month', mt.month,
            'category_name', mt.category_name,
            'amount', mt.monthly_expenses
        )
    ) FILTER (WHERE mt.month IS NOT NULL) as monthly_trend
FROM yearly_totals yt
LEFT JOIN top_categories tc ON 
    tc.currency = yt.currency 
    AND tc.rank <= 5
LEFT JOIN monthly_trend mt ON 
    mt.currency = yt.currency
GROUP BY 
    yt.currency, 
    yt.total_income, 
    yt.total_expenses, 
    yt.total_savings;
