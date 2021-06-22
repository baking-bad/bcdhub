create materialized view if not exists head_stats AS
select network, count(*) as value, 'calls_count' as stats_type from operations where entrypoint != '' group by network
union all
select network, count(*) as value, 'contracts_count' as stats_type from contracts group by network
union all
select network, count(distinct(hash)) as value, 'unique_contracts_count' as stats_type from contracts group by network
union all
select network, count(*) as value, 'fa_count' as stats_type from contracts where (tags & 448) > 0 group by network;

create unique index if not exists head_stats_idx on head_stats(network, stats_type);