create or replace 
	function update_contract_stats_on_delete_operation () returns trigger as	
	$$
	declare
   		ts timestamp;
	begin
		if exists (select id from contracts c2 where address = old.destination and network = old.network) THEN
			select "timestamp" into ts from operations 
				where (destination = old.destination or source = old.destination) and network = old.network 
				order by level limit 1;

			update contracts set tx_count = tx_count - 1, last_action = ts
				where address = old.destination;
		end if;
		return old;
	end;
	$$
language 'plpgsql';

drop trigger if exists contract_stats_on_delete_operation ON operations;

create trigger contract_stats_on_delete_operation
	before delete on operations
	for each row execute 
	procedure update_contract_stats_on_delete_operation();