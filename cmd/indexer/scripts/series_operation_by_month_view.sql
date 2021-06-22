create materialized view if not exists series_operation_by_month_%s AS
with f as (
        select generate_series(
        date_trunc('month', date '2018-06-25'),
        date_trunc('month', now()),
        '1 month'::interval
        ) as val
)
select
        extract(epoch from f.val),
        count(*) as value
from f
left join operations on date_trunc('month', operations.timestamp) = f.val where ((network = %d) and (entrypoint is not null and entrypoint != '') and (status = 1))
group by 1
order by date_part;

create unique index if not exists series_operation_by_month_%s_idx on series_operation_by_month_%s(date_part);